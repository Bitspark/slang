package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

var printPorts bool

func main() {
	flag.BoolVar(&printPorts, "print-ports", false, "display port def")

	if len(os.Args) < 2 {
		fmt.Println("USAGE: slang [OPTIONS] SLANGFILE.slang.json")
		fmt.Println("OPTIONS:")
		flag.PrintDefaults()
		return
	}

	flag.Parse()

	slFile, err := readSlangFile(flag.Arg(0))

	if err != nil {
		log.Fatal(err)
	}

	if printPorts {
		if err := printPortDef(slFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := run(slFile); err != nil {
		log.Fatal(err)
	}

}

func readSlangFile(slFilePath string) (*core.SlangFileDef, error) {
	var slFile core.SlangFileDef

	b, err := ioutil.ReadFile(slFilePath)

	if err != nil {
		return &slFile, fmt.Errorf("could not read operator file: %s", slFilePath)
	}
	err = json.Unmarshal(b, &slFile)
	return &slFile, err
}

func printPortDef(slFile *core.SlangFileDef) error {
	mainBpId := slFile.Main
	var opDef *core.OperatorDef
	for _, bp := range slFile.Blueprints {
		if mainBpId == bp.Id {
			opDef = &bp
		}
	}

	if opDef == nil {
		return fmt.Errorf("unknown blueprint: %s", mainBpId)
	}

	fmt.Printf("Ports:\n")
	fmt.Printf("\tIn:\n\t\t%s\n", jsonString(opDef.ServiceDefs["main"].In))
	fmt.Printf("\tOut:\n\t\t%s\n", jsonString(opDef.ServiceDefs["main"].Out))

	return nil
}

func run(slFile *core.SlangFileDef) error {

	errors := make(chan error, 1)
	done := make(chan bool, 1)
	portcfgs := make(chan map[string]string, 1)

	cmdr := api.NewCommander(":0")

	go func() {
		err := cmdr.Begin(func(c api.Commands) error {
			var msg string
			var err error

			msg, err = c.Hello()

			if err != nil {
				return err
			}

			if msg == "" {
				msg, err = c.Init(jsonString(slFile))
			} else {
				msg, err = c.PrtCfg()
			}

			if err != nil {
				return err
			}

			var pcfg map[string]string
			if err = json.Unmarshal([]byte(msg), &pcfg); err != nil {
				return err
			}

			portcfgs <- pcfg
			return nil
		})

		if err != nil {
			errors <- err
		}
	}()

	cmd := exec.Command("slangr", "--aggr-in", "--aggr-out", "--mgnt-addr", fmt.Sprintf("%s", cmdr.Addr()))

	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			<-done
		}
		done <- true
	}()

	var portcfg map[string]string
	select {
	case err = <-errors:
		return err
	case portcfg = <-portcfgs:
		break
	}

	pconn := api.NewPortConnHandler(portcfg)
	if err := pconn.ConnectTo("(", pushToRnr); err != nil {
		return err
	}
	if err := pconn.ConnectTo(")", pullFromRnr); err != nil {
		return err
	}

	<-done
	return nil
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
