package core

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	TYPE_GENERIC   = iota
	TYPE_PRIMITIVE = iota
	TYPE_TRIGGER   = iota
	TYPE_NUMBER    = iota
	TYPE_STRING    = iota
	TYPE_BINARY    = iota
	TYPE_BOOLEAN   = iota
	TYPE_STREAM    = iota
	TYPE_MAP       = iota
)

const (
	_             = iota
	DIRECTION_IN  = iota
	DIRECTION_OUT = iota
)

var CHANNEL_SIZE = 1 << 15
var CHANNEL_DYNAMIC = false

type BOS struct {
	src *Port
}

type EOS struct {
	src *Port
}

type PH struct {
	t string
}

var PHSingle = &PH{"..."}
var PHMultiple = &PH{"[...]"}

type Port struct {
	operator  *Operator
	service   *Service
	delegate  *Delegate
	dests     map[*Port]bool
	src       *Port
	strSrc    *Port
	direction int

	itemType int

	parStr *Port
	parMap *Port

	sub  *Port
	subs map[string]*Port

	buf    chan interface{}
	mutex  sync.Mutex
	closed bool
}

// Makes a new port.
func NewPort(srv *Service, del *Delegate, def TypeDef, dir int) (*Port, error) {
	if !def.valid {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	if dir != DIRECTION_IN && dir != DIRECTION_OUT {
		return nil, errors.New("wrong direction")
	}

	p := &Port{}
	p.strSrc = p
	p.direction = dir
	if srv != nil {
		p.operator = srv.operator
	} else if del != nil {
		p.operator = del.operator
	}
	p.service = srv
	p.delegate = del
	p.dests = make(map[*Port]bool)

	var err error
	switch def.Type {
	case "map":
		p.itemType = TYPE_MAP
		p.subs = make(map[string]*Port)
		for k, e := range def.Map {
			p.subs[k], err = NewPort(srv, del, *e, dir)
			if err != nil {
				return nil, err
			}
			p.subs[k].parStr = p.parStr
			p.subs[k].parMap = p
		}
	case "stream":
		p.itemType = TYPE_STREAM
		p.sub, err = NewPort(srv, del, *def.Stream, dir)
		if err != nil {
			return nil, err
		}
		setParentStreams(p.sub, p)
	case "primitive":
		p.itemType = TYPE_PRIMITIVE
	case "trigger":
		p.itemType = TYPE_TRIGGER
	case "number":
		p.itemType = TYPE_NUMBER
	case "string":
		p.itemType = TYPE_STRING
	case "binary":
		p.itemType = TYPE_BINARY
	case "boolean":
		p.itemType = TYPE_BOOLEAN
	}

	if p.PrimitiveType() && dir == DIRECTION_IN && p.operator != nil && p.operator.function != nil {
		p.buf = make(chan interface{}, CHANNEL_SIZE)
	}

	return p, nil
}

// Returns the type of the port.
func (p *Port) Type() int {
	return p.itemType
}

// Returns the direction of the port.
func (p *Port) Direction() int {
	return p.direction
}

// Returns the parent stream port. Port must be of type stream.
func (p *Port) ParentStream() *Port {
	return p.parStr
}

// Returns the operator this port is attached to
func (p *Port) Operator() *Operator {
	return p.operator
}

// Returns the operator this port is attached to
func (p *Port) Delegate() *Delegate {
	return p.delegate
}

// Returns the subport with the according name of this port. Port must be of type map.
func (p *Port) Map(name string) *Port {
	port, _ := p.subs[name]
	return port
}

func (p *Port) Lock() {
	p.mutex.Lock()
}

func (p *Port) Unlock() {
	p.mutex.Unlock()
}

// Returns the length of the map ports
func (p *Port) MapLength() int {
	return len(p.subs)
}

// Returns all map entry names of this port
func (p *Port) MapEntryNames() []string {
	entries := []string{}
	for entry := range p.subs {
		entries = append(entries, entry)
	}
	return entries
}

// Returns the substream port of this port. Port must be of type stream.
func (p *Port) Stream() *Port {
	return p.sub
}

// Connects this port with port p.
func (p *Port) Connect(q *Port) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%s", e))
		}
	}()

	if q.src != nil {
		if q.src == p {
			return nil
		}
		return fmt.Errorf("%s -> %s: already connected", p.String(), q.String())
	}

	if q.itemType == TYPE_PRIMITIVE {
		return p.connect(q, true)
	}

	if q.itemType == TYPE_TRIGGER {
		if p.itemType == TYPE_MAP {
			for _, sub := range p.subs {
				return sub.Connect(q)
			}
			return fmt.Errorf("%s -> %s: trigger connected with empty map", p.Name(), q.Name())
		}
		return p.connect(q, true)
	}

	if p.itemType != TYPE_PRIMITIVE && p.itemType != q.itemType || p.itemType == TYPE_PRIMITIVE && !q.PrimitiveType() {
		return fmt.Errorf("%s -> %s: types don't match - %d != %d", p.Name(), q.Name(), p.itemType, q.itemType)
	}

	if p.PrimitiveType() {
		return p.connect(q, true)
	}

	if p.itemType == TYPE_MAP {
		if len(p.subs) != len(q.subs) {
			return fmt.Errorf("%s -> %s: maps are incompatible - unequal lengths %d and %d", p.Name(), q.Name(), len(p.subs), len(q.subs))
		}

		for k, pe := range p.subs {
			qe, ok := q.subs[k]
			if !ok {
				return fmt.Errorf("%s -> %s: maps are incompatible - %s not present", p.Name(), q.Name(), k)
			}

			err := pe.Connect(qe)

			if err != nil {
				return err
			}
		}

		return nil
	}

	if p.itemType == TYPE_STREAM {
		if q.sub == nil {
			return fmt.Errorf("%s -> %s: streams are incompatible - no sub present", p.Name(), q.Name())
		}

		return p.sub.Connect(q.sub)
	}

	if p.itemType == TYPE_GENERIC {
		return fmt.Errorf("%s -> %s: cannot connect generic type", p.Name(), q.Name())
	}

	return fmt.Errorf("%s -> %s: unknown type", p.Name(), q.Name())
}

// Disconnects this port from port q.
func (p *Port) Disconnect(q *Port) error {
	if !p.Connected(q) {
		return errors.New("not connected")
	}
	q.src = nil
	delete(p.dests, q)
	return nil
}

// Returns true if p is connected with q.
func (p *Port) Connected(q *Port) bool {
	if b, ok := p.dests[q]; ok {
		return b && q.src == p
	}
	return false
}

// Opens the port by opening all channels
func (p *Port) Open() {
	if !p.closed {
		return
	}

	p.closed = false

	if p.buf != nil {
		p.buf = make(chan interface{}, CHANNEL_SIZE)
	}

	if p.sub != nil {
		p.sub.Open()
	}

	if len(p.subs) > 0 {
		for _, sub := range p.subs {
			sub.Open()
		}
	}
}

// Closes the port by closing all channels
func (p *Port) Close() {
	if p.closed {
		return
	}
	p.Lock()
	p.closed = true
	if p.buf != nil {
		close(p.buf)
	}
	p.Unlock()

	if p.sub != nil {
		p.sub.Close()
	}

	if len(p.subs) > 0 {
		for _, sub := range p.subs {
			sub.Close()
		}
	}
}

func (p *Port) StreamSource() *Port {
	return p.strSrc
}

func (p *Port) SetStreamSource(srcStr *Port) {
	p.strSrc = srcStr
}

// Removes this port and redirects connections.
func (p *Port) Merge() bool {
	merged := false

	if p.src != nil {
		for dest := range p.dests {
			p.src.dests[dest] = true
			dest.src = p.src
			dest.strSrc = p.strSrc
			merged = true
		}
		p.src.Disconnect(p)
	} else if len(p.dests) != 0 {
		for dest := range p.dests {
			if dest.itemType != TYPE_TRIGGER {
				continue
			}

			p.strSrc.dests[dest] = true
			dest.strSrc = p.strSrc
			merged = true
		}
	}

	if p.sub != nil {
		merged = p.sub.Merge() || merged
	}

	for _, sub := range p.subs {
		merged = sub.Merge() || merged
	}

	return merged
}

// Checks if this in port is connected completely
func (p *Port) DirectlyConnected() error {
	if p.direction != DIRECTION_IN {
		return errors.New("can only check in ports")
	}

	if p.PrimitiveType() {
		if p.src == nil {
			return errors.New(p.Name() + " not connected")
		}
		if p.src.direction != DIRECTION_OUT && p.src.operator.name != "" {
			return errors.New(p.Name() + " only out ports can be connected with in ports")
		}
		if p.src.src != nil {
			return errors.New(p.Name() + " has connected source " + p.src.Name() + ": " + p.src.src.Name())
		}
		for dest := range p.src.dests {
			if dest == p {
				return nil
			}
		}
		return errors.New(p.Name() + " not connected back from " + p.src.Name())
	}

	if p.sub != nil {
		if err := p.sub.DirectlyConnected(); err != nil {
			return err
		}
		return nil
	}

	for _, sub := range p.subs {
		if err := sub.DirectlyConnected(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Port) assertChannelSpace() {
	c := cap(p.buf)
	if len(p.buf) > c/2 {
		newChan := make(chan interface{}, 2*c)
		p.Lock()
		for {
			select {
			case i := <-p.buf:
				newChan <- i
			default:
				goto end
			}
		}
	end:
		p.buf = newChan
		p.Unlock()
	}
}

func (p *Port) WalkPrimitivePorts(handle func(p *Port)) {
	if p.PrimitiveType() {
		handle(p)
	}

	if p.Stream() != nil {
		p.Stream().WalkPrimitivePorts(handle)
	}

	for _, pname := range p.MapEntryNames() {
		p.Map(pname).WalkPrimitivePorts(handle)
	}
}
func (p *Port) Closed() bool {
	return p.closed
}

// Push an item to this port.
func (p *Port) Push(item interface{}) {
	p.Lock()
	if p.closed {
		p.Unlock()
		return
	}
	p.Unlock()

	if p.buf != nil {
		if CHANNEL_DYNAMIC {
			p.assertChannelSpace()

			p.Lock()
			p.buf <- item
			p.Unlock()
		} else {
			p.buf <- item
		}
	}

	for dest := range p.dests {
		if dest.Type() == TYPE_TRIGGER || p.PrimitiveType() {
			dest.Push(item)
		}
	}

	if p.PrimitiveType() {
		return
	}

	if p.itemType == TYPE_MAP {
		m, ok := item.(map[string]interface{})

		if !ok {
			for _, sub := range p.subs {
				sub.Push(item)
			}
			return
		}

		for k, i := range m {
			if sub, ok := p.subs[k]; ok {
				sub.Push(i)
			}
		}
		return
	}

	if p.itemType == TYPE_STREAM {
		items, ok := item.([]interface{})
		if !ok {
			p.sub.Push(item)
			return
		}

		p.PushNoTriggerBOS()
		for _, i := range items {
			p.sub.Push(i)
		}
		p.PushEOS()
	}
}

func (p *Port) PushNoTriggerBOS() {
	p.sub.Push(BOS{p.strSrc})
}

func (p *Port) PushBOS() {
	// For triggers, we need to push right here
	for dest := range p.dests {
		if dest.Type() == TYPE_TRIGGER {
			dest.Push(nil)
		}
	}

	p.sub.Push(BOS{p.strSrc})
}

func (p *Port) PushEOS() {
	p.sub.Push(EOS{p.strSrc})
}

// Pull an item from this port
func (p *Port) Pull() interface{} {
	if p.itemType == TYPE_GENERIC {
		panic("cannot pull from generic")
	}

	if p.buf != nil {
		if CHANNEL_DYNAMIC {
			for {
				p.Lock()
				select {
				case i := <-p.buf:
					p.Unlock()
					return i
				default:
					p.Unlock()
				}
				time.Sleep(1 * time.Millisecond)
			}
		} else {
			i, ok := <-p.buf
			if !ok {
				// the channel was activly closed
				// but we still send the zero value
				// we recievied - this solves a strange
				// race condition detected with `go test -race ...`
				return i
			}
			return i
		}
	}

	if p.PrimitiveType() {
		panic("no buffer")
	}

	if p.itemType == TYPE_MAP {
		var mi interface{}
		itemMap := make(map[string]interface{})

		for k, sub := range p.subs {
			i := sub.Pull()

			if i == PHMultiple {
				mi = PHMultiple
				continue
			}
			if bos, ok := i.(BOS); ok {
				mi = bos
				continue
			}
			if eos, ok := i.(EOS); ok {
				mi = eos
				continue
			}
			itemMap[k] = i
		}

		if mi != nil {
			return mi
		}
		return itemMap
	}

	if p.itemType == TYPE_STREAM {
		i := p.sub.Pull()

		if !p.OwnBOS(i) {
			return i
		}

		items := []interface{}{}

		for {
			i := p.sub.Pull()

			if p.OwnEOS(i) {
				return items
			}

			items = append(items, i)
		}
	}

	panic("unknown type")
}

// Pull a float
func (p *Port) PullFloat64() (float64, interface{}) {
	item := p.Pull()
	if f, ok := item.(float64); ok {
		return f, nil
	}
	if i, ok := item.(int); ok {
		return float64(i), nil
	}
	return 0, item
}

// Pull an int
func (p *Port) PullInt() (int, interface{}) {
	item := p.Pull()
	if i, ok := item.(int); ok {
		return i, nil
	}
	if f, ok := item.(float64); ok {
		return int(f), nil
	}
	return 0, item
}

// Pull a string
func (p *Port) PullString() (string, interface{}) {
	item := p.Pull()
	if s, ok := item.(string); ok {
		return s, nil
	}
	return "", item
}

// Pull an boolean
func (p *Port) PullBoolean() (bool, interface{}) {
	item := p.Pull()
	if b, ok := item.(bool); ok {
		return b, nil
	}
	return false, item
}

// Pull a binary object
func (p *Port) PullBinary() (Binary, interface{}) {
	item := p.Pull()
	if b, ok := item.(Binary); ok {
		return b, nil
	}
	return nil, item
}

func (p *Port) PullBOS() bool {
	i := p.sub.Pull()
	if !p.OwnBOS(i) {
		panic(p.operator.Name() + ": expected own BOS: " + i.(BOS).src.String() + " != " + p.strSrc.String())
	}
	return true
}

func (p *Port) PullEOS() bool {
	if !p.OwnEOS(p.sub.Pull()) {
		panic("expected own EOS")
	}
	return true
}

// Similar to Port.Pull but will return nil when there is no item after timeout
func (p *Port) Poll() interface{} {
	if p.buf == nil {
		panic("no buffer")
	}

	if len(p.buf) == 0 {
		time.Sleep(200 * time.Millisecond)
		if len(p.buf) == 0 {
			return nil
		}
	}

	var i interface{}
	if CHANNEL_DYNAMIC {
		p.Lock()
		i = <-p.buf
		p.Unlock()
	} else {
		i = <-p.buf
	}

	return i
}

func (p *Port) NewBOS() BOS {
	return BOS{p.strSrc}
}

func (p *Port) NewEOS() EOS {
	return EOS{p.strSrc}
}

func (p *Port) OwnBOS(i interface{}) bool {
	if bos, ok := i.(BOS); ok {
		// (bos.src == p) is only the case if i has directly been pushed into p
		if (p.strSrc != nil && bos.src == p.strSrc) || bos.src == p {
			return true
		}
	}
	return false
}

func (p *Port) OwnEOS(i interface{}) bool {
	if eos, ok := i.(EOS); ok {
		// (eos.src == p) is only the case if i has directly been pushed into p
		if (p.strSrc != nil && eos.src == p.strSrc) || eos.src == p {
			return true
		}
	}
	return false
}

// Name returns a name generated from operator, directions and port.
func (p *Port) Name() string {
	if p == nil {
		return "<nil>"
	}

	return p.String()
}

func (p *Port) Bufferize() {
	if p.buf != nil {
		return
	}

	if p.PrimitiveType() {
		p.buf = make(chan interface{}, CHANNEL_SIZE)
	} else if p.itemType == TYPE_MAP {
		for _, sub := range p.subs {
			sub.Bufferize()
		}
	} else if p.itemType == TYPE_STREAM {
		p.sub.Bufferize()
	}
}

// PRIVATE METHODS

func setParentStreams(p *Port, parent *Port) {
	p.parStr = parent

	if p.itemType == TYPE_MAP {
		for _, sub := range p.subs {
			setParentStreams(sub, parent)
		}
	}
}

func (p *Port) wire(q *Port, original bool) {
	if original {
		p.dests[q] = true
		q.src = p
	}
	q.strSrc = p.strSrc
	if q.operator != nil && q.operator.connectFunc != nil {
		q.operator.connectFunc(q.operator, q, p)
	}
}

func (p *Port) connect(q *Port, original bool) error {
	if p.direction == DIRECTION_IN {
		if q.direction == DIRECTION_IN {
			if p.operator != q.operator.Parent() {
				return errors.New("wrong operator nesting")
			}

			p.wire(q, original)

			if q.parStr == nil {
				q.operator.basePort = p.parStr
			} else if p.parStr != nil {
				p.parStr.connect(q.parStr, false)
			}
		} else {
			if p.operator != q.operator {
				return errors.New("wrong operator nesting")
			}

			p.wire(q, original)

			if p.parStr != nil && q.parStr != nil {
				p.parStr.connect(q.parStr, false)
			}
		}
	} else {
		if q.direction == DIRECTION_IN {
			if p.operator.Parent() != q.operator.Parent() {
				return errors.New("wrong operator nesting")
			}

			p.wire(q, original)

			if q.parStr == nil {
				if p.parStr != nil {
					q.operator.basePort = p.parStr
				} else {
					q.operator.basePort = p.operator.basePort
				}
			} else {
				if p.parStr != nil {
					p.parStr.connect(q.parStr, false)
				} else if p.operator.basePort != nil {
					p.operator.basePort.connect(q.parStr, false)
				}
			}
		} else {
			if p.operator.Parent() != q.operator {
				return errors.New("wrong operator nesting")
			}

			p.wire(q, original)

			if p.parStr != nil {
				if q.parStr != nil {
					// TODO: Investigate this
					// We probably don't need this for the first stage of connecting, only when connecting flattened
					// operators this is necessary
					p.parStr.connect(q.parStr, false)
				}
			} else if p.operator.basePort != nil {
				if q.parStr != nil {
					// TODO: See above
					p.operator.basePort.connect(q.parStr, false)
				}
			}
		}
	}

	return nil
}

func (p *Port) PrimitiveType() bool {
	return p.itemType == TYPE_PRIMITIVE ||
		p.itemType == TYPE_TRIGGER ||
		p.itemType == TYPE_NUMBER ||
		p.itemType == TYPE_STRING ||
		p.itemType == TYPE_BINARY ||
		p.itemType == TYPE_BOOLEAN
}

func (p *Port) TriggerType() bool {
	return p.itemType == TYPE_TRIGGER
}

func (p *Port) MapType() bool {
	return p.itemType == TYPE_MAP
}

func (p *Port) Define() TypeDef {
	var def TypeDef

	switch p.itemType {
	case TYPE_PRIMITIVE:
		def.Type = "primitive"
	case TYPE_TRIGGER:
		def.Type = "trigger"
	case TYPE_STRING:
		def.Type = "string"
	case TYPE_NUMBER:
		def.Type = "number"
	case TYPE_BOOLEAN:
		def.Type = "boolean"
	case TYPE_BINARY:
		def.Type = "binary"
	case TYPE_GENERIC:
		def.Type = "generic"
	case TYPE_STREAM:
		def.Type = "stream"
		subDef := p.sub.Define()
		def.Stream = &subDef
	case TYPE_MAP:
		def.Type = "map"
		def.Map = make(map[string]*TypeDef)
		for k, sub := range p.subs {
			subDef := sub.Define()
			def.Map[k] = &subDef
		}
	}

	return def
}

func (p *Port) stringify() string {
	if p.parMap != nil {
		parMapStr := p.parMap.stringify()
		entryStr := ""
		for k, ps := range p.parMap.subs {
			if ps == p {
				entryStr = k
				break
			}
		}
		if parMapStr == "" {
			return entryStr
		}
		return parMapStr + "." + entryStr
	}

	if p.parStr != nil {
		parStrStr := p.parStr.stringify()
		if parStrStr == "" {
			return "~"
		}
		return parStrStr + ".~"
	}

	return ""
}

func (p *Port) String() string {
	opStr := ""
	if p.operator != nil {
		opStr = p.operator.name
	}
	if p.service != nil {
		if p.service.name != MAIN_SERVICE {
			opStr = p.service.name + "@" + opStr
		}
	} else if p.delegate != nil {
		opStr = opStr + "." + p.delegate.name
	}

	portStr := p.stringify()

	if p.direction == DIRECTION_IN {
		return portStr + "(" + opStr
	} else if p.direction == DIRECTION_OUT {
		return opStr + ")" + portStr
	}

	return ""
}

func (p *Port) defineConnections(def *Blueprint) {
	portStr := p.String()

	if def.Connections[portStr] == nil {
		def.Connections[portStr] = make([]string, 0)
		for dst := range p.dests {
			def.Connections[portStr] = append(def.Connections[portStr], dst.String())
			dst.operator.defineConnections(def)
		}
	}

	if p.sub != nil {
		p.sub.defineConnections(def)
	}

	for _, sub := range p.subs {
		sub.defineConnections(def)
	}
}

func (ph *PH) MarshalJSON() ([]byte, error) {
	return []byte("\"@PH " + ph.t + "\""), nil
}
