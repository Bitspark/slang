package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/capability"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/log"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/thoas/go-funk"
)

var SupportedRunModes = []string{"process", "httpPost"}

// safeModeFromEnv reads SLANG_SAFE_MODE, which leaves shell execution and file
// writes unregistered. The hosted runner sets it. It is an environment variable
// rather than a flag because older binaries ignore an unknown variable, so the
// runner can set it before every deployed engine release understands it. An
// unreadable value is an error, never a silent return to the unsafe default.
func safeModeFromEnv(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	safe, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid SLANG_SAFE_MODE %q: must be true or false", value)
	}
	return safe, nil
}

// capabilitiesFromEnv reads SLANG_HTTP_CAPABILITY_SOCKET. When it names a capability
// host's socket, HTTP requests go through that host instead of this machine's
// network; the hosted runner, which has no network, sets it.
func capabilitiesFromEnv(socket string) elem.Capabilities {
	caps := elem.LocalCapabilities()
	if socket != "" {
		caps.HTTP = &http.Client{Timeout: caps.HTTP.Timeout, Transport: capability.NewHTTPTransport(socket)}
	}
	return caps
}

// listen opens bind, a TCP address or "unix:" followed by a socket path. A socket
// left behind by an earlier run is replaced; any other file at that path is kept.
func listen(bind string) (net.Listener, error) {
	socket := strings.TrimPrefix(bind, "unix:")
	if socket == bind {
		return net.Listen("tcp", bind)
	}
	if info, err := os.Lstat(socket); err == nil && info.Mode()&os.ModeSocket != 0 {
		os.Remove(socket)
	}
	return net.Listen("unix", socket)
}

func main() {
	runMode := flag.String("mode", SupportedRunModes[0], fmt.Sprintf("Choose run mode for operator: %s", SupportedRunModes))
	bind := flag.String("bind", "localhost:0", "Address httpPost listens on: host:port, or unix: followed by a socket path")
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("slang OPTIONS SLANG_BUNDLE")
		flag.PrintDefaults()
	}

	// Check cmd args

	// Expect slang file as 1st arg
	slangBundlePath := flag.Arg(0)
	if slangBundlePath == "" {
		log.Fatal("missing slang bundle file")
	}

	// Expect supported runmode
	if !funk.ContainsString(SupportedRunModes, *runMode) {
		log.Fatalf("invalid run mode: %s must be one of following %s", *runMode, SupportedRunModes)
	}

	// Read in slang file
	slBundle, err := readSlangBundleJSON(slangBundlePath)
	if err != nil {
		log.Fatal(err)
	}

	// Init elementary operators
	safeMode, err := safeModeFromEnv(os.Getenv("SLANG_SAFE_MODE"))
	if err != nil {
		log.Fatal(err)
	}
	elem.SafeMode = safeMode
	elem.Init()

	// Parse and Build blueprint
	operator, err := api.BuildOperatorWith(slBundle, capabilitiesFromEnv(os.Getenv("SLANG_HTTP_CAPABILITY_SOCKET")))
	if err != nil {
		log.Fatal(err)
	}

	log.SetBlueprint(operator.Id(), operator.Name())

	// Run
	if err := run(operator, *runMode, *bind); err != nil {
		log.Fatal(err)
	}

}

func readSlangBundleJSON(slBundlePath string) (*core.SlangBundle, error) {
	slBundleContent, err := ioutil.ReadFile(slBundlePath)

	if err != nil {
		return nil, err
	}

	var slFile core.SlangBundle
	err = json.Unmarshal([]byte(slBundleContent), &slFile)
	return &slFile, err
}

func run(operator *core.Operator, mode string, bind string) error {
	if operator.Main() == nil {
		return errors.New("blueprint has no main service")
	}
	if err := operator.Main().Out().FullyConnected(); err != nil {
		return err
	}
	// Handle SIGTERM (CTRL-C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	switch mode {
	case "process":
		go func() {
			runProcess(operator)
			quit <- syscall.SIGQUIT
		}()
	case "httpPost":
		go func() {
			runHttpPost(operator, bind)
			quit <- syscall.SIGQUIT
		}()
	default:
		return fmt.Errorf("run mode not supported: %s", mode)
	}

	for {
		select {
		case <-quit:
			return nil
		case <-time.After(5 * time.Second):
			log.Ping()
		}
	}
}

func runProcess(operator *core.Operator) {
	// Expect to read from stdin
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if fi.Mode()&(os.ModeNamedPipe|os.ModeSocket) == 0 {
		log.Fatal("slang command is intended to work with pipes\nUsage: data-src | slang")
	}

	operator.Main().Out().Bufferize()
	operator.Start()

	/*
		if isQuasiTrigger(operator.Main().In()) {
			operator.Main().In().Push(true)
		}
	*/

	incoming := make(chan interface{})
	outgoing := make(chan interface{})
	stopped := false
	// expecting to read newline delimited json (ndjson) from stdin
	jdeco := json.NewDecoder(os.Stdin)

	// Read from stdin
	go func() {
	loop:
		for jdeco.More() {
			var jval interface{}
			if err := jdeco.Decode(&jval); err != nil {
				// as soon as decode error decoder cannot continue to read stream
				// without break this line will be passed infinitly
				log.Error("json decode error: ", err)
				break loop
			}
			jval = core.CleanValue(jval)
			incoming <- jval
		}
		stopped = true
	}()

	// Write to stdout
	go func() {
		var jval interface{}
		jenco := json.NewEncoder(os.Stdout)

	loop:
		for !stopped {
			jval = <-outgoing
			if err := jenco.Encode(jval); err != nil {
				log.Error("json encode error: ", err)
				break loop
			}
		}
		operator.Stop()
	}()

	go func() {
	loop:
		for {
			jval := <-incoming
			operator.Main().In().Push(jval)

			p := operator.Main().Out()
			if p.Closed() {
				break loop
			}

			item, err := p.Receive()
			if err != nil {
				return
			}
			outgoing <- item
		}
	}()

	operator.WaitForStop()
}

func runHttpPost(operator *core.Operator, bind string) {
	inDef := operator.Main().In().Define()

	r := mux.NewRouter()
	r.
		Methods("POST").
		HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			var incoming interface{}

			err := json.NewDecoder(req.Body).Decode(&incoming)
			switch {
			// We do not have a POST-Body but we could still serve a result
			// for the case when the `In` is a trigger.
			case err == io.EOF:
				if isQuasiTrigger(operator.Main().In()) {
					operator.Main().In().Push(true)
					outgoing, err := operator.Main().Out().Receive()
					if err != nil {
						responseWithError(resp, err, http.StatusServiceUnavailable)
						return
					}
					responseWithOk(resp, outgoing)
				} else {
					responseWithError(resp, errors.New("missing data"), http.StatusBadRequest)
				}
			// We have an error while decoding the response
			case err != nil:
				responseWithError(resp, err, http.StatusBadRequest)

			// Everything is fine, validate the values and pass it to the running operator
			default:
				incoming = core.CleanValue(incoming)
				if err := inDef.VerifyData(incoming); err != nil {
					responseWithError(resp, err, http.StatusBadRequest)
					return
				}
				operator.Main().In().Push(incoming)

				p := operator.Main().Out()
				if p.Closed() {
					return
				}

				outgoing, err := p.Receive()
				if err != nil {
					responseWithError(resp, err, http.StatusServiceUnavailable)
					return
				}
				responseWithOk(resp, outgoing)
			}

		})

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"POST"},
	}).Handler(r)

	operator.Main().Out().Bufferize()
	operator.Start()
	listener, err := listen(bind)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(listener, handler))
}

func isQuasiTrigger(p *core.Port) bool {
	// port is quasi a trigger,
	// when it actually is a trigger port or
	// it is a map with in total one sub-port of trigger type
	return p.TriggerType() || p.MapType() && p.MapLength() == 1 && p.Map(p.MapEntryNames()[0]).TriggerType()
}

func responseWithError(w http.ResponseWriter, err error, status int) {
	log.Error(err)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(err.Error()); err != nil {
		log.Fatal(err)
	}
}

func responseWithOk(w http.ResponseWriter, m interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Fatal(err)
	}
}
