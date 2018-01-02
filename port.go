package main

import (
	"errors"
	"fmt"
	"reflect"
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
	if def == nil || reflect.ValueOf(def).IsNil() {
		return nil, errors.New("definition is nil")
	}

	p := &Port{}
	p.direction = dir
	p.operator = o
	p.dests = make(map[*Port]bool)

	itemType, ok := def["type"]

	if !ok {
		return nil, errors.New("type missing")
	}

	var err error
	switch itemType {
	case "map":
		p.itemType = TYPE_MAP
		p.subs = make(map[string]*Port)
		me, ok := def["map"]
		if !ok {
			return nil, errors.New("map missing")
		}
		m, ok := me.(map[string]interface{})
		if !ok {
			return nil, errors.New("map malformed")
		}
		for k, ee := range m {
			e, ok := ee.(map[string]interface{})
			if !ok {
				return nil, errors.New("entry malformed")
			}
			p.subs[k], err = MakePort(o, e, dir)
			if err != nil {
				return nil, err
			}
			p.subs[k].parStr = p.parStr
			p.subs[k].parMap = p
		}
	case "stream":
		p.itemType = TYPE_STREAM
		se, ok := def["stream"]
		if !ok {
			return nil, errors.New("stream missing")
		}
		s, ok := se.(map[string]interface{})
		if !ok {
			return nil, errors.New("stream malformed")
		}
		p.sub, err = MakePort(o, s, dir)
		if err != nil {
			return nil, err
		}
		p.sub.parStr = p
	default:
		p.itemType = TYPE_ANY

		if dir == DIRECTION_IN && o != nil && o.function != nil {
			p.buf = make(chan interface{}, 100)
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

	if p.itemType != q.itemType {
		return errors.New("types don't match")
	}

	switch p.itemType {
	case TYPE_ANY:
		return p.connect(q)
	case TYPE_MAP:

		if len(p.subs) != len(q.subs) {
			return errors.New("maps are incompatible: Unequal lengths")
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

	if p.itemType == TYPE_ANY {
		panic("no buffer")
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

	if p.itemType == TYPE_MAP {
		m, ok := item.(map[string]interface{})

		if !ok {
			for _, sub := range p.subs {
				sub.Push(m)
			}
			return
		}

		for k, i := range m {
			p.subs[k].Push(i)
		}
		return
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

	panic("unknown type")
}

// Name returns a name generated from operator, directions and port.
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
