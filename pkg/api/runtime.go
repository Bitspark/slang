package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Bitspark/go-funk"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Commander interface {
	Begin(action func(c Commands) error) error
	Addr() string
}

type Worker interface {
	Begin(new func() Commands) error
	Addr() string
}

type Commands interface {
	Hello() (string, error)
	Init(a string) (string, error)
	PrtCfg() (string, error)
	Action() error
}

type cmdr struct {
	addr string
	ln   net.Listener
}

type wrkr struct {
	addr string
}

type cmdrCmdsImpl struct {
	wr     *bufio.Writer
	rd     *bufio.Reader
	action func(c Commands) error
}

func NewCommander(addr string) Commander {
	ln, _ := net.Listen("tcp", addr)
	return &cmdr{ln.Addr().String(), ln}
}

func NewWorker(addr string) Worker {
	return &wrkr{addr}
}

func (m *cmdr) Addr() string {
	return m.addr
}

func (m *wrkr) Addr() string {
	return m.addr
}

func (m *cmdr) Begin(action func(c Commands) error) error {
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

			c := &cmdrCmdsImpl{bufio.NewWriter(conn), bufio.NewReader(conn), action}
			go func() {
				if err := c.Action(); err != nil {
					errors <- err
				}
			}()
		}
	}()

	return <-errors
}

func (m *wrkr) Begin(newWorkerCmds func() Commands) error {
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

func (m *wrkr) dispatch(conn net.Conn, c Commands, errors chan error, wg *sync.WaitGroup) {
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	defer wg.Done()

	for {
		var msg string
		var rmsg string
		var err error

		fmt.Println("--> dispatch")
		msg, err = Rdbuf(rd)

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

		err = Wrbuf(wr, rmsg)
		if err != nil {
			errors <- err
			break
		}

	}

}

func Wrbuf(wr *bufio.Writer, msg string) error {
	msg = strings.TrimSpace(msg)
	if _, err := wr.WriteString(msg + "\n"); err != nil {
		return err
	}
	return wr.Flush()
}

func Rdbuf(rd *bufio.Reader) (string, error) {
	msg, err := rd.ReadString('\n')

	if err != nil {
		return msg, err
	}

	msg = strings.TrimSpace(msg)
	return msg, nil
}

func JsonWrbuf(wr *bufio.Writer, j interface{}) error {
	msg, err := json.Marshal(j)

	if err != nil {
		return err
	}

	if err = Wrbuf(wr, string(msg)); err != nil {
		return err
	}

	return nil
}

func JsonRdbuf(rd *bufio.Reader) (interface{}, error) {
	msg, err := Rdbuf(rd)

	if err != nil && err != io.EOF {
		return nil, err
	}

	var j interface{}

	if len(msg) > 0 {
		if err = json.Unmarshal([]byte(msg), &j); err != nil {
			return nil, err
		}
	}

	return j, nil
}

func (c *cmdrCmdsImpl) Action() error {
	return c.action(c)
}

func (c *cmdrCmdsImpl) Hello() (string, error) {
	fmt.Println("---> write hello")
	err := Wrbuf(c.wr, "/hello")

	fmt.Println("---> ", err)

	if err != nil {
		return "", err
	}

	return Rdbuf(c.rd)
}

func (c *cmdrCmdsImpl) Init(a string) (string, error) {
	if err := Wrbuf(c.wr, "/init "+a); err != nil {
		return "", err
	}
	return Rdbuf(c.rd)
}

func (c *cmdrCmdsImpl) PrtCfg() (string, error) {
	if err := Wrbuf(c.wr, "/ports"); err != nil {
		return "", err
	}
	return Rdbuf(c.rd)
}

type PortConnHandler interface {
	ListPortRefs() []string
	ConnectTo(p string, hndl func(c net.Conn)) error
}

type prtScktMap struct {
	pmap map[string]string
}

func NewPortConnHandler(pmap map[string]string) PortConnHandler {
	return &prtScktMap{pmap}
}

func (ps *prtScktMap) ListPortRefs() []string {
	return funk.Keys(ps.pmap).([]string)
}

func (ps *prtScktMap) ConnectTo(p string, hndl func(c net.Conn)) error {
	addr, ok := ps.pmap[p]

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		for {
			if conn, err := net.Dial("tcp", addr); err == nil {
				go func() {
					hndl(conn)
					wg.Done()
				}()

				wg.Wait()
				wg.Add(1)
			}
			time.Sleep(100 * time.Millisecond)
		}

	}()

	if !ok {
		return fmt.Errorf("unknown port: %s", p)
	}
	return nil
}
