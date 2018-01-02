package main

import (
	"errors"
)

const (
	TYPE_ANY    = iota
	TYPE_STREAM = iota
	TYPE_MAP    = iota
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

// PUBLIC METHODS

// Makes a new port.
func MakePort(o *Operator, def map[string]interface{}, dir int) (*Port, error) {
	var err error

	p := &Port{}
	p.direction = dir
	p.operator = o
	p.dests = make(map[*Port]bool)

	switch def["type"] {
	case "map":
		p.itemType = TYPE_MAP
		p.subs = make(map[string]*Port)
		for k, e := range def["map"].(map[string]interface{}) {
			p.subs[k], err = MakePort(o, e.(map[string]interface{}), dir)
			if err != nil {
				return nil, err
			}
			p.subs[k].parStr = p.parStr
			p.subs[k].parMap = p
		}
	case "stream":
		p.itemType = TYPE_STREAM
		p.sub, err = MakePort(o, def["stream"].(map[string]interface{}), dir)
		if err != nil {
			return nil, err
		}
		p.sub.parStr = p
	default:
		p.itemType = TYPE_ANY

		if dir == DIRECTION_IN && o != nil && o.function != nil {
			p.buf = make(chan interface{})
		}
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
	if p.itemType != TYPE_ANY || q.itemType != TYPE_ANY {
		return errors.New("can only connect primitives")
	}

	return p.connect(q)
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
	}

	if p.itemType == TYPE_ANY {
		for dest := range p.dests {
			dest.Push(item)
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

	if p.itemType == TYPE_ANY {
		panic("no buffer")
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

// Returns a name generated from operator, directions and port.
func (p *Port) Name() string {
	var name string

	switch p.itemType {
	case TYPE_STREAM:
		name = "STREAM"
	default:
		name = "ANY"
	}

	if p.parStr == nil {
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

	return p.parStr.Name() + "_" + name
}

// PRIVATE METHODS

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
				q.operator.basePort = p.parStr
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
