package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/log"
)

var SupportedRunModes = []string{"process"}

func main() {
	runMode := flag.String("mode", SupportedRunModes[0], fmt.Sprintf("Choose run mode for operator: %s", SupportedRunModes))
	slangFilePath := flag.String("file", "", fmt.Sprintf("Path to slangFile"))
	flag.Parse()

	if !funk.ContainsString(SupportedRunModes, *runMode) {
		log.Fatalf("invalid run mode: %s must be one of following %s", runMode, SupportedRunModes)
	}

	var slangFileReader io.Reader
	var err error

	if *slangFilePath != "" {
		slangFileReader, err = os.Open(*slangFilePath)
	} else {
		err = checkStdin()
		slangFileReader = os.Stdin
	}

	if err != nil {
		log.Fatalf("provide slangFile via stdin or via flag -file=slangFile")
	}

	slFileBufReader := bufio.NewReader(slangFileReader)
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

func readSlangFile(reader *bufio.Reader) (*core.SlangFileDef, error) {
	slFileContent, err := api.Rdbuf(reader)

	if err != nil && err != io.EOF {
		return nil, err
	}

	var slFile core.SlangFileDef
	err = json.Unmarshal([]byte(slFileContent), &slFile)
	return &slFile, err
}

func run(operator *core.Operator, mode string) error {
	switch mode {
	case "process":
		runProcess(operator)
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

func isQuasiTrigger(p *core.Port) bool {
	// port is quasi a trigger,
	// when it actually is a trigger port or
	// it is a map with in total one sub-port of trigger type
	return p.TriggerType() || p.MapType() && p.MapLength() == 1 && p.Map(p.MapEntryNames()[0]).TriggerType()
}

func runProcess(operator *core.Operator) {
	operator.Main().Out().Bufferize()
	operator.Start()
	log.Print("started")

	if isQuasiTrigger(operator.Main().In()) {
		log.Warn("auto trigger")
		operator.Main().In().Push(true)
	}
}
