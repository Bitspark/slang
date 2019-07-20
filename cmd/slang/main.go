package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/log"
	"github.com/gorilla/mux"
)

var SupportedRunModes = []string{"process", "httpPost"}

func main() {
	runMode := flag.String("mode", SupportedRunModes[0], fmt.Sprintf("Choose run mode for operator: %s", SupportedRunModes))
	slangBundlePath := flag.String("file", "", fmt.Sprintf("Path to slangBundle"))
	flag.Parse()

	if !funk.ContainsString(SupportedRunModes, *runMode) {
		log.Fatalf("invalid run mode: %s must be one of following %s", runMode, SupportedRunModes)
	}

	var slangBundleReader io.Reader
	var err error

	if *slangBundlePath != "" {
		slangBundleReader, err = os.Open(*slangBundlePath)
	} else {
		err = checkStdin()
		slangBundleReader = os.Stdin
	}

	if err != nil {
		log.Fatalf("provide slangBundle via stdin or via flag -file=slangBundle")
	}

	slFileBufReader := bufio.NewReader(slangBundleReader)
	slFile, err := readSlangFile(slFileBufReader)

	if err != nil {
		log.Fatal(err)
	}

	operator, err := api.BuildOperator(slFile)

	if err != nil {
		log.Fatal(err)
	}

	log.SetOperator(operator.Id(), operator.Name())

	if err := run(operator, *runMode); err != nil {
		log.Fatal(err)
	}

}

func checkStdin() error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return fmt.Errorf("empty stdin")
	}

	return nil
}

func readSlangFile(reader *bufio.Reader) (*core.SlangBundle, error) {
	slFileContent, err := api.Rdbuf(reader)

	if err != nil && err != io.EOF {
		return nil, err
	}

	var slFile core.SlangBundle
	err = json.Unmarshal([]byte(slFileContent), &slFile)
	return &slFile, err
}

func run(operator *core.Operator, mode string) error {
	switch mode {
	case "process":
		runProcess(operator)
	case "httpPost":
		runHttpPost(operator)
	default:
		log.Fatal("invalid or not implemented run mode: %s")
	}

	// Handle SIGTERM (CTRL-C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

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
	operator.Main().Out().Bufferize()
	operator.Start()
	log.Print("started")

	if isQuasiTrigger(operator.Main().In()) {
		operator.Main().In().Push(true)
	}
}

func runHttpPost(operator *core.Operator) {
	inDef := operator.Main().In().Define()

	r := mux.NewRouter()
	r.
		Methods("POST").
		HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			var incoming interface{}

			if err := json.NewDecoder(req.Body).Decode(&incoming); err != nil {
				responseWithError(resp, err, http.StatusBadRequest)
				return
			}

			incoming = core.CleanValue(incoming)
			if err := inDef.VerifyData(incoming); err != nil {
				responseWithError(resp, err, http.StatusBadRequest)
				return
			}

			operator.Main().In().Push(incoming)

			outgoing := operator.Main().Out().Pull()

			responseWithOk(resp, outgoing)
		})

	operator.Main().Out().Bufferize()
	operator.Start()
	log.Print("started")

	go func() {
		log.Fatal(http.ListenAndServe("localhost:0", r))
	}()
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
