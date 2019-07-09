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
	runMode := *flag.String("mode", SupportedRunModes[0], fmt.Sprintf("Choose run mode for operator: %s", SupportedRunModes))
	flag.Parse()

	if !funk.ContainsString(SupportedRunModes, runMode) {
		log.Fatalf("invalid run mode: %s must be one of following %s", runMode, SupportedRunModes)
	}

	if err := checkStdin(); err != nil {
		log.Fatal(err)
	}

	slFileReader := bufio.NewReader(os.Stdin)
	slFile, err := readSlangFile(slFileReader)

	if err != nil {
		log.Fatal(err)
	}

	operator, err := api.BuildOperator(slFile)

	if err != nil {
		log.Fatal(err)
	}

	log.SetOperator(operator.Id(), operator.Name())

	if err := run(operator, runMode); err != nil {
		log.Fatal(err)
	}

}

func checkStdin() error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return fmt.Errorf("needs slangFile via stdin")
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
		operator.Main().Out().Bufferize()
		operator.Start()

		if operator.Main().In().TriggerType() {
			log.Warnf("is a trigger input: %s", operator.Main().In())
			operator.Main().In().Push(true)
		} else {
			log.Warnf("is NOT a trigger input: %s", operator.Main().In())
		}
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
