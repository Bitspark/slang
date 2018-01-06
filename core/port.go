package core

import (
	"errors"
	"fmt"
)

const (
	TYPE_ANY       = iota
	TYPE_PRIMITIVE = iota
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
	portDef   PortDef
	operator  *Operator
	dests     map[*Port]bool
	src       *Port
	direction int

	itemType int

	parStr *Port
	parMap *Port

	sub  *Port
	subs map[string]*Port
	any  string

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
	p.portDef = def
	p.direction = dir
	p.operator = o
	p.dests = make(map[*Port]bool)

	var err error
	switch def.Type {
	case "map":
		p.itemType = TYPE_MAP
		p.subs = make(map[string]*Port)
		for k, e := range def.Map {
			p.subs[k], err = NewPort(o, e, dir)
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
	case "number":
		p.itemType = TYPE_NUMBER
	case "string":
		p.itemType = TYPE_STRING
	case "boolean":
		p.itemType = TYPE_BOOLEAN
	case "any":
		p.itemType = TYPE_ANY
		p.any = def.Any
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

// Returns the subport with the according name of this port. Port must be of type map.
func (p *Port) Map(name string) *Port {
	if p.itemType != TYPE_MAP {
		panic("not a map")
	}
	port, _ := p.subs[name]
	return port
}

// Returns the substream port of this port. Port must be of type stream.
func (p *Port) Stream() *Port {
	if p.itemType != TYPE_STREAM {
		panic("not a stream")
	}
	return p.sub
}

// Returns the any identifier of this port. Port must be of type any.
func (p *Port) Any(name string) string {
	if p.itemType != TYPE_ANY {
		panic("not an any")
	}
	return p.any
}

// Connects this port with port p.
func (p *Port) Connect(q *Port) error {
	// Types don't match
	if p.itemType != q.itemType && p.itemType != TYPE_ANY && q.itemType != TYPE_ANY &&
		!(p.itemType == TYPE_PRIMITIVE && q.primitive() || p.primitive() && q.itemType == TYPE_PRIMITIVE) {
		return errors.New(fmt.Sprintf("types don't match: %d != %d", p.itemType, q.itemType))
	}

	propagateForwards := false
	if p.itemType != TYPE_ANY && q.itemType == TYPE_ANY {
		if err := q.operator.SpecifyAny(q.any, p.portDef); err != nil {
			return err
		}
		propagateForwards = true
	}

	propagateBackwards := false
	if p.itemType == TYPE_ANY && q.itemType != TYPE_ANY {
		if err := p.operator.SpecifyAny(p.any, q.portDef); err != nil {
			return err
		}
		propagateBackwards = true
	}

	if p.primitive() || q.itemType == TYPE_ANY || p.itemType == TYPE_ANY {
		err := p.connect(q)

		if err != nil {
			return err
		}

		if propagateForwards {
			if err := q.operator.PropagateAnys(); err != nil {
				return err
			}
		}

		if propagateBackwards {
			if err := p.operator.PropagateAnys(); err != nil {
				return err
			}
		}

		return nil
	}

	if p.itemType == TYPE_MAP {
		if len(p.subs) != len(q.subs) {
			return errors.New("maps are incompatible: unequal lengths")
		}

		for k, pe := range p.subs {
			qe, ok := q.subs[k]
			if !ok {
				return errors.New(fmt.Sprintf("maps are incompatible: %s not present", k))
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
			return errors.New("streams are incompatible: no sub present")
		}

		return p.sub.Connect(q.sub)
	}

	return errors.New("could not connect")
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
	if p.itemType == TYPE_ANY {
		panic("cannot push to any")
	}

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
		return
	}

	panic(fmt.Sprintf("pushing to unknown type %d", p.itemType))
}

func (p *Port) PushBOS() {
	p.Push(BOS{p})
}

func (p *Port) PushEOS() {
	p.Push(EOS{p})
}

// Pull an item from this port.
func (p *Port) Pull() interface{} {
	if p.itemType == TYPE_ANY {
		panic("cannot pull from any")
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

	panic(fmt.Sprintf("pulling from unknown type %d", p.itemType))
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
		if bos.src == p.src || bos.src == p {
			return true
		}
	}
	return false
}

func (p *Port) OwnEOS(i interface{}) bool {
	if eos, ok := i.(EOS); ok {
		// (eos.src == p) is only the case if i has directly been pushed into p
		if eos.src == p.src || eos.src == p {
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

	switch p.itemType {
	case TYPE_ANY:
		name = "ANY"
	case TYPE_PRIMITIVE:
		name = "PRIMITIVE"
	case TYPE_NUMBER:
		name = "NUMBER"
	case TYPE_STRING:
		name = "STRING"
	case TYPE_BOOLEAN:
		name = "BOOLEAN"
	case TYPE_MAP:
		name = "MAP"
	case TYPE_STREAM:
		name = "STREAM"
	}

	if p.parMap != nil {
		return p.parMap.Name() + "_" + name
	}

	if p.parStr != nil {
		return p.parStr.Name() + "_" + name
	}

	if p.direction == DIRECTION_IN {
		if p.operator != nil {
			return p.operator.name + ":IN_" + name
		} else {
			return "IN_" + name
		}
	} else {
		if p.operator != nil {
			return p.operator.name + ":OUT_" + name
		} else {
			return "OUT_" + name
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
	return p.itemType == TYPE_PRIMITIVE || p.itemType == TYPE_NUMBER || p.itemType == TYPE_STRING || p.itemType == TYPE_BOOLEAN
}
