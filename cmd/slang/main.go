package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Bitspark/slang/pkg/api"
	"io/ioutil"
	"log"
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
	portcfgs := make(chan interface{}, 1)

	cmdr := api.NewCommanderConnHandler(":0")
	err = cmdr.Begin(func(c api.Cmds) error {
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

		var pcfg interface{}
		if err = json.Unmarshal([]byte(msg), &pcfg); err != nil {
			fmt.Println("---> failed", msg, err)
			return err
		}

		fmt.Println("---> ports", pcfg, err)

		portcfgs <- pcfg
		return nil
	})

	if err != nil {
		return err
	}

	cmd := exec.Command("slangr", "--mgmnt-addr", cmdr.Addr())

	err = cmd.Start()
	if err != nil {
		return err
	}

	fmt.Println("---> Slang runner started")

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Println("--> err", err)
			errors <- err
		}
		done <- true
	}()

	var portcfg interface{}
	select {
	case err = <-errors:
		return err
	case portcfg = <-portcfgs:
		break
	}

	fmt.Println("--->", portcfg)
	log.Printf("Waiting for command to finish...")
	<-done
	return nil
}

/*
func obtainPortCfg(ln net.Listener) (interface{}, error) {
	fmt.Println("---> waiting", ln.Addr().String())
	conn, err := ln.Accept()
	fmt.Println("---> listen", ln.Addr().String())

	if err != nil {
		return nil, err
	}

	_, err = bufio.NewWriter(conn).WriteString("\n")

	if err != nil {
		return nil, err
	}

	msg, err := bufio.NewReader(conn).ReadString('\n')

	fmt.Println("---> read", msg)

	if err != nil {
		return nil, err
	}

	var pcfg interface{}
	err = json.Unmarshal([]byte(msg), &pcfg)

	return pcfg, err
}

func readPipeDecodeJSON(rd *bufio.Reader, wrerr *bufio.Writer) (interface{}, bool) {
	var idat interface{}
	for {
		text, err := rd.ReadString('\n')

		if err == io.EOF {
			return nil, true
		}

		text = strings.TrimSpace(text)

		if len(text) == 0 {
			return nil, false
		}

		err = json.Unmarshal([]byte(text), &idat)

		if err != nil {
			wrerr.WriteString(err.Error() + "\n")
			wrerr.Flush()
			continue
		}

		return idat, false
	}
}

*/
