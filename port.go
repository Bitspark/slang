package slang

import (
	"errors"
	"fmt"
)

const (
	TYPE_ANY     = iota
	TYPE_NUMBER  = iota
	TYPE_STRING  = iota
	TYPE_BOOLEAN = iota
	TYPE_STREAM  = iota
	TYPE_MAP     = iota
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

type portDef struct {
	Type   string             `json:"type"`
	Stream *portDef           `json:"stream"`
	Map    map[string]portDef `json:"map"`
	valid  bool
}

// PUBLIC METHODS

// Makes a new port.
func MakePort(o *Operator, def portDef, dir int) (*Port, error) {
	if !def.valid {
		err := def.validate()
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
			p.subs[k], err = MakePort(o, e, dir)
			if err != nil {
				return nil, err
			}
			p.subs[k].parStr = p.parStr
			p.subs[k].parMap = p
		}
	case "stream":
		p.itemType = TYPE_STREAM
		p.sub, err = MakePort(o, *def.Stream, dir)
		if err != nil {
			return nil, err
		}
		setParentStreams(p.sub, p)
	case "number":
		p.itemType = TYPE_NUMBER
	case "string":
		p.itemType = TYPE_STRING
	case "boolean":
		p.itemType = TYPE_BOOLEAN
	case "any":
		p.itemType = TYPE_ANY
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
func (p *Port) Port(name string) *Port {
	return p.subs[name]
}

// Returns the substream port of this port. Port must be of type stream.
func (p *Port) Stream() *Port {
	return p.sub
}

// Connects this port with port p.
func (p *Port) Connect(q *Port) error {
	if p.itemType != TYPE_ANY && q.itemType != TYPE_ANY && p.itemType != q.itemType {
		return errors.New(fmt.Sprintf("types don't match: %d != %d", p.itemType, q.itemType))
	}

	if p.primitive() {
		return p.connect(q)
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

	return errors.New("can only connect primitives and maps")
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

		p.sub.Push(BOS{p})
		for _, i := range items {
			p.sub.Push(i)
		}
		p.sub.Push(EOS{p})
	}
}

// Pull an item from this port.
func (p *Port) Pull() interface{} {
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

		if bos, ok := i.(BOS); ok {
			if bos.src != p.src {
				return i
			}
		} else {
			return i
		}

		items := []interface{}{}

		for true {
			i := p.sub.Pull()

			if eos, ok := i.(EOS); ok {
				if eos.src == p.src {
					return items
				}
			}

			items = append(items, i)
		}
	}

	panic("unknown type")
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

// PRIVATE METHODS

func (d *portDef) validate() error {
	validTypes := []string{"any", "number", "string", "boolean", "stream", "map"}
	found := false
	for _, t := range validTypes {
		if t == d.Type {
			found = true
			break
		}
	}
	if !found {
		return errors.New("unknown type")
	}

	if d.Type == "stream" {
		if d.Stream == nil {
			return errors.New("stream missing")
		}
		return d.Stream.validate()
	} else if d.Type == "map" {
		if len(d.Map) == 0 {
			return errors.New("map missing or empty")
		}
		for _, e := range d.Map {
			err := e.validate()
			if err != nil {
				return err
			}
		}
	}

	d.valid = true
	return nil
}

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
	return p.itemType == TYPE_ANY || p.itemType == TYPE_NUMBER || p.itemType == TYPE_STRING || p.itemType == TYPE_BOOLEAN
}
