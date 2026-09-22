package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/thoas/go-funk"
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
			conn, err := ln.Accept()

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
			if conn, err := net.Dial("tcp", m.addr); err == nil {
				c := newWorkerCmds()
				go m.dispatch(conn, c, errors, &wg)
				go func() {
					if err := c.Action(); err != nil {
						errors <- err
					}
				}()
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

		msg, err = Rdbuf(rd)

		if err != nil {
			if err != io.EOF {
				errors <- err
			}
			break
		}

		s := strings.SplitN(msg, " ", 2)

		switch s[0] {
		case "/hello":
			rmsg, err = c.Hello()
		case "/init":
			rmsg, err = c.Init(s[1])
		case "/ports":
			rmsg, err = c.PrtCfg()
		default:
			continue
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
	err := Wrbuf(c.wr, "/hello")

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
	ConnectTo(p string, hndl func(c net.Conn) bool) error
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

func (ps *prtScktMap) ConnectTo(p string, hndl func(c net.Conn) bool) error {
	addr, ok := ps.pmap[p]

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		for {
			if conn, err := net.Dial("tcp", addr); err == nil {
				reconn := true

				go func() {
					reconn = hndl(conn)
					wg.Done()
				}()

				wg.Wait()

				if !reconn {
					return
				}

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
