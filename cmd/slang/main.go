package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/log"
)

func main() {
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

	if err := run(operator); err != nil {
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

func run(operator *core.Operator) error {
	operator.Main().Out().Bufferize()
	operator.Start()

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
