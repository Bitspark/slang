package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Bitspark/slang/pkg/api"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()

	slangFile := flag.Arg(0)

	fmt.Println("---> Args", slangFile)

	if err := run(slangFile); err != nil {
		log.Fatal(err)
	}

}

func run(slangFile string) error {
	b, err := ioutil.ReadFile(slangFile)

	if err != nil {
		return fmt.Errorf("could not read operator file: %s", slangFile)
	}

	errors := make(chan error, 1)
	done := make(chan bool, 1)
	portcfgs := make(chan map[string]string, 1)

	fmt.Println("-> new commander")
	cmdr := api.NewCommander(":0")
	fmt.Println("-> begin")

	go func() {
		err := cmdr.Begin(func(c api.Commands) error {
			var msg string
			var err error

			fmt.Println("--> /", msg)

			msg, err = c.Hello()

			fmt.Println("--> /hello", msg, err)

			if err != nil {
				return err
			}

			if msg == "" {
				fmt.Println("--> requires init")
				msg, err = c.Init(string(b))
				fmt.Println("--> /init", msg, err)
			} else {
				msg, err = c.PrtCfg()
				fmt.Println("--> /ports", msg, err)
			}

			fmt.Println("---> ready", msg, err)

			if err != nil {
				return err
			}

			var pcfg map[string]string
			if err = json.Unmarshal([]byte(msg), &pcfg); err != nil {
				fmt.Println("---> failed", msg, err)
				return err
			}

			fmt.Println("---> ports", pcfg, err)

			portcfgs <- pcfg
			return nil
		})

		if err != nil {
			errors <- err
		}
	}()

	fmt.Println("about to spawn slang runner:", fmt.Sprintf("--mgnt-addr \"%s\"", cmdr.Addr()))
	cmd := exec.Command("slangr", "--aggr-in", "--aggr-out", "--mgnt-addr", fmt.Sprintf("%s", cmdr.Addr()))
	fmt.Println("===>", cmd.Args)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Start()
	if err != nil {
		return err
	}

	fmt.Println("---> Slang runner started")

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Println("--> err", err)
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

	fmt.Println("--->", portcfg)

	pconn := api.NewPortConnHandler(portcfg)
	if err := pconn.ConnectTo("(", pushToRnr); err != nil {
		return err
	}
	if err := pconn.ConnectTo(")", pullFromRnr); err != nil {
		return err
	}

	log.Printf("Waiting for command to finish...")
	<-done
	return nil
}

func wrerr(err error) {
	wrerr := bufio.NewWriter(os.Stderr)
	wrerr.WriteString(err.Error() + "\n")
	wrerr.Flush()
}

func pushToRnr(connRnr net.Conn) {
	fmt.Println("push to", connRnr)
	stdin := bufio.NewReader(os.Stdin)
	wrRnr := bufio.NewWriter(connRnr)

	defer connRnr.Close()

	for {
		m, err := api.Rdbuf(stdin)

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

}

func pullFromRnr(connRnr net.Conn) {
	rdRnr := bufio.NewReader(connRnr)
	stdout := bufio.NewWriter(os.Stdin)

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
			break
		}
	}
}
