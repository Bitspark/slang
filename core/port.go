package core

import (
	"errors"
	"fmt"
)

const (
	TYPE_GENERIC   = iota
	TYPE_PRIMITIVE = iota
	TYPE_TRIGGER   = iota
	TYPE_NUMBER    = iota
	TYPE_STRING    = iota
	TYPE_BOOLEAN   = iota
	TYPE_STREAM    = iota
	TYPE_MAP       = iota
)

const (
	_             = iota
	DIRECTION_IN  = iota
	DIRECTION_OUT = iota
)

type BOS struct {
	src *Port
}

type EOS struct {
	src *Port
}

type Port struct {
	operator  *Operator
	dests     map[*Port]bool
	src       *Port
	direction int

	itemType int

	parStr *Port
	parMap *Port

	sub  *Port
	subs map[string]*Port

	buf chan interface{}
}

// Makes a new port.
func NewPort(o *Operator, def PortDef, dir int) (*Port, error) {
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
	p.direction = dir
	p.operator = o
	p.dests = make(map[*Port]bool)

	var err error
	switch def.Type {
	case "map":
		p.itemType = TYPE_MAP
		p.subs = make(map[string]*Port)
		for k, e := range def.Map {
			p.subs[k], err = NewPort(o, *e, dir)
			if err != nil {
				return nil, err
			}
			p.subs[k].parStr = p.parStr
			p.subs[k].parMap = p
		}
	case "stream":
		p.itemType = TYPE_STREAM
		p.sub, err = NewPort(o, *def.Stream, dir)
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
	case "boolean":
		p.itemType = TYPE_BOOLEAN
	}

	if p.primitive() && dir == DIRECTION_IN && o != nil && o.function != nil {
		p.buf = make(chan interface{}, 100)
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

// Returns the subport with the according name of this port. Port must be of type map.
func (p *Port) Map(name string) *Port {
	port, _ := p.subs[name]
	return port
}

// Returns the substream port of this port. Port must be of type stream.
func (p *Port) Stream() *Port {
	return p.sub
}

// Connects this port with port p.
func (p *Port) Connect(q *Port) error {
	if p.itemType != TYPE_PRIMITIVE && q.itemType != TYPE_PRIMITIVE && p.itemType != q.itemType {
		return fmt.Errorf("%s -> %s: types don't match - %d != %d", p.Name(), q.Name(), p.itemType, q.itemType)
	}

	if p.primitive() {
		return p.connect(q)
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

// Removes this port and redirects connections.
func (p *Port) Merge() bool {
	merged := false

	if p.src != nil {
		for dest := range p.dests {
			p.src.dests[dest] = true
			dest.src = p.src
			merged = true
		}
		p.src.Disconnect(p)
	}

	if p.sub != nil {
		merged = p.sub.Merge() || merged
	}

	return merged
}

// Push an item to this port.
func (p *Port) Push(item interface{}) {
	if p.buf != nil {
		p.buf <- item
		return
	}

	if p.primitive() {
		for dest := range p.dests {
			dest.Push(item)
		}
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
			p.subs[k].Push(i)
		}
		return
	}

	if p.itemType == TYPE_STREAM {
		items, ok := item.([]interface{})
		if !ok {
			p.sub.Push(item)
			return
		}

		p.PushBOS()
		for _, i := range items {
			p.sub.Push(i)
		}
		p.PushEOS()
	}
}

func (p *Port) PushBOS() {
	p.Push(BOS{p})
}

func (p *Port) PushEOS() {
	p.Push(EOS{p})
}

// Pull an item from this port.
func (p *Port) Pull() interface{} {
	if p.itemType == TYPE_GENERIC {
		panic("cannot pull from generic")
	}

	if p.buf != nil {
		return <-p.buf
	}

	if p.primitive() {
		panic("no buffer")
	}

	if p.itemType == TYPE_MAP {
		var mi interface{}
		itemMap := make(map[string]interface{})

		for k, sub := range p.subs {
			i := sub.Pull()

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

		for true {
			i := p.sub.Pull()

			if p.OwnEOS(i) {
				return items
			}

			items = append(items, i)
		}
	}

	panic("unknown type")
}

func (p *Port) NewBOS() BOS {
	return BOS{p}
}

func (p *Port) NewEOS() EOS {
	return EOS{p}
}

func (p *Port) OwnBOS(i interface{}) bool {
	if bos, ok := i.(BOS); ok {
		// (bos.src == p) is only the case if i has directly been pushed into p
		if (p.src != nil && bos.src == p.src) || bos.src == p {
			return true
		}
	}
	return false
}

func (p *Port) OwnEOS(i interface{}) bool {
	if eos, ok := i.(EOS); ok {
		// (eos.src == p) is only the case if i has directly been pushed into p
		if (p.src != nil && eos.src == p.src) || eos.src == p {
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

	var name string

	if p.parMap != nil {
		for pn, pt := range p.parMap.subs {
			if pt == p {
				name = "[" + pn + "]"
			}
		}
	}

	switch p.itemType {
	case TYPE_GENERIC:
		name += "_GENERIC"
	case TYPE_PRIMITIVE:
		name += "_PRIMITIVE"
	case TYPE_TRIGGER:
		name += "_TRIGGER"
	case TYPE_NUMBER:
		name += "_NUMBER"
	case TYPE_STRING:
		name += "_STRING"
	case TYPE_BOOLEAN:
		name += "_BOOLEAN"
	case TYPE_MAP:
		name += "_MAP"
	case TYPE_STREAM:
		name += "_STREAM"
	}

	if p.parMap != nil {
		return p.parMap.Name() + name
	}

	if p.parStr != nil {
		return p.parStr.Name() + name
	}

	if p.direction == DIRECTION_IN {
		if p.operator != nil {
			return p.operator.name + ":IN" + name
		} else {
			return "IN_" + name
		}
	} else {
		if p.operator != nil {
			return p.operator.name + ":OUT" + name
		} else {
			return "OUT" + name
		}
	}
}

func (p *Port) Bufferize() chan interface{} {
	if p.buf == nil {
		p.buf = make(chan interface{}, 100)
	}
	return p.buf
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

func (p *Port) connect(q *Port) error {
	if p.direction == DIRECTION_IN {
		if q.direction == DIRECTION_IN {
			if p.operator != q.operator.Parent() {
				return errors.New("wrong operator nesting")
			}

			p.dests[q] = true
			q.src = p

			if q.parStr == nil {
				q.operator.basePort = p.parStr
			} else if p.parStr != nil {
				p.parStr.connect(q.parStr)
			}
		} else {
			if p.operator != q.operator {
				return errors.New("wrong operator nesting")
			}

			p.dests[q] = true
			q.src = p

			if p.parStr != nil && q.parStr != nil {
				p.parStr.connect(q.parStr)
			}
		}
	} else {
		if q.direction == DIRECTION_IN {
			if p.operator.Parent() != q.operator.Parent() {
				return errors.New("wrong operator nesting")
			}

			p.dests[q] = true
			q.src = p

			if q.parStr == nil {
				if p.parStr != nil {
					q.operator.basePort = p.parStr
				} else {
					q.operator.basePort = p.operator.basePort
				}
			} else {
				if p.parStr != nil {
					p.parStr.connect(q.parStr)
				} else if p.operator.basePort != nil {
					p.operator.basePort.connect(q.parStr)
				}
			}
		} else {
			if p.operator.Parent() != q.operator {
				return errors.New("wrong operator nesting")
			}

			p.dests[q] = true
			q.src = p

			if p.parStr != nil {
				p.parStr.connect(q.parStr)
			} else if p.operator.basePort != nil {
				p.operator.basePort.connect(q.parStr)
			}
		}
	}

	return nil
}

func (p *Port) primitive() bool {
	return p.itemType == TYPE_PRIMITIVE ||
		p.itemType == TYPE_TRIGGER ||
		p.itemType == TYPE_NUMBER ||
		p.itemType == TYPE_STRING ||
		p.itemType == TYPE_BOOLEAN
}
