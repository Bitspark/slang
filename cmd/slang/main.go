package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
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

func printPortDef(slFile *core.SlangFileDef) error {
	mainBpId := slFile.Main
	var blueprint *core.Blueprint
	for _, bp := range slFile.Blueprints {
		bp := bp
		if mainBpId == bp.Id {
			blueprint = &bp
		}
	}

	if blueprint == nil {
		return fmt.Errorf("unknown blueprint: %s", mainBpId)
	}

	fmt.Printf("Ports:\n")
	fmt.Printf("\tIn:\n\t\t%s\n", jsonString(blueprint.ServiceDefs["main"].In))
	fmt.Printf("\tOut:\n\t\t%s\n", jsonString(blueprint.ServiceDefs["main"].Out))

	return nil
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

func jsonString(j interface{}) string {
	jb, _ := json.Marshal(j)
	return string(jb)
}

func wrerr(err error) {
	wrerr := bufio.NewWriter(os.Stderr)
	wrerr.WriteString(err.Error() + "\n")
	wrerr.Flush()
}

func pushToRnr(connRnr net.Conn) bool {
	stdin := bufio.NewReader(os.Stdin)
	wrRnr := bufio.NewWriter(connRnr)

	defer connRnr.Close()

	for {
		m, err := api.Rdbuf(stdin)

		time.Sleep(1 * time.Second)

		if err == io.EOF {
			break
		}

		if err != nil {
			wrerr(err)
			continue
		}

		if err := api.Wrbuf(wrRnr, m); err != nil {
			break
		}
	}

	return false
}

func pullFromRnr(connRnr net.Conn) bool {
	rdRnr := bufio.NewReader(connRnr)
	stdout := bufio.NewWriter(os.Stdout)

	defer connRnr.Close()

	for {
		m, err := api.Rdbuf(rdRnr)

		if err == io.EOF {
			break
		}

		if err != nil {
			wrerr(err)
			continue
		}

		if err := api.Wrbuf(stdout, m); err != nil {
			wrerr(err)
			break
		}
	}
	return false
}
