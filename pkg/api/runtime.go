package api

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type CommanderConnHandler interface {
	Begin(action func(c Cmds) error) error
	Addr() string
}

type WorkerConnHandler interface {
	Begin(new func() Cmds) error
	Addr() string
}

type Cmds interface {
	Hello() (string, error)
	Init(a string) (string, error)
	PrtCfg() (string, error)
	Action() error
}

type cmdConnHandlr struct {
	addr string
	ln   net.Listener
}

type wkrConnHandlr struct {
	addr string
}

type cmds struct {
	wr     *bufio.Writer
	rd     *bufio.Reader
	action func(c Cmds) error
}

func NewCommanderConnHandler(addr string) CommanderConnHandler {
	ln, _ := net.Listen("tcp", addr)
	return &cmdConnHandlr{ln.Addr().String(), ln}
}

func NewWorkerConnHandler(addr string) WorkerConnHandler {
	return &wkrConnHandlr{addr}
}

func (m *cmdConnHandlr) Addr() string {
	return m.addr
}

func (m *wkrConnHandlr) Addr() string {
	return m.addr
}

func (m *cmdConnHandlr) Begin(action func(c Cmds) error) error {
	ln := m.ln

	errors := make(chan error, 1)

	go func() {
		for {
			fmt.Println("---> waiting", ln.Addr().String())
			conn, err := ln.Accept()
			fmt.Println("---> listen", ln.Addr().String())

			if err != nil {
				errors <- err
			}

			c := &cmds{bufio.NewWriter(conn), bufio.NewReader(conn), action}
			go func() {
				if err := c.Action(); err != nil {
					errors <- err
				}
			}()
		}
	}()

	return <-errors
}

func (m *wkrConnHandlr) Begin(newWorkerCmds func() Cmds) error {
	errors := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		for {
			fmt.Println("--> lets go")
			if conn, err := net.Dial("tcp", m.addr); err == nil {
				fmt.Println("--> connected")
				c := newWorkerCmds()
				go m.dispatch(conn, c, errors, &wg)
				go func() {
					if err := c.Action(); err != nil {
						errors <- err
					}
				}()
				fmt.Println("--> waiting")
				wg.Wait()
				wg.Add(1)
			}
			time.Sleep(100 * time.Millisecond)
		}

	}()
	return <-errors
}

func (m *wkrConnHandlr) dispatch(conn net.Conn, c Cmds, errors chan error, wg *sync.WaitGroup) {
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	defer wg.Done()

	for {
		var msg string
		var rmsg string
		var err error

		fmt.Println("--> dispatch")
		msg, err = readBuf(rd)

		if err != nil {
			errors <- err
			break
		}

		s := strings.SplitN(msg, " ", 2)

		fmt.Println("---> incoming cmd", s[0])

		switch s[0] {
		case "/hello":
			rmsg, err = c.Hello()
		case "/init":
			rmsg, err = c.Init(s[1])
		case "/ports":
			rmsg, err = c.PrtCfg()
		default:
			fmt.Println("---> unkwon command", s[0])
		}

		if err != nil {
			errors <- err
			break
		}

		err = writeBuf(wr, rmsg)
		if err != nil {
			errors <- err
			break
		}

	}

}

func writeBuf(wr *bufio.Writer, msg string) error {
	msg = strings.TrimSpace(msg)
	if _, err := wr.WriteString(msg + "\n"); err != nil {
		return err
	}
	return wr.Flush()
}

func readBuf(rd *bufio.Reader) (string, error) {
	msg, err := rd.ReadString('\n')
	if err != nil {
		return msg, err
	}
	msg = strings.TrimSpace(msg)
	return msg, nil
}

func (c *cmds) Action() error {
	return c.action(c)
}

func (c *cmds) Hello() (string, error) {
	fmt.Println("---> write hello")
	err := writeBuf(c.wr, "/hello")

	fmt.Println("---> ", err)

	if err != nil {
		return "", err
	}

	return readBuf(c.rd)
}

func (c *cmds) Init(a string) (string, error) {
	if err := writeBuf(c.wr, "/init "+a); err != nil {
		return "", err
	}
	return readBuf(c.rd)
}

func (c *cmds) PrtCfg() (string, error) {
	if err := writeBuf(c.wr, "/ports"); err != nil {
		return "", err
	}
	return readBuf(c.rd)
}
