package core

import (
	"errors"
	"fmt"
	"strings"
)

type InstanceDef struct {
	Operator   string                 `json:"operator"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
	Ports      map[string]PortDef     `json:"ports"`

	valid       bool
	operatorDef OperatorDef
}

type OperatorDef struct {
	In          PortDef             `json:"in"`
	Out         PortDef             `json:"out"`
	Operators   []*InstanceDef      `json:"operators"`
	Connections map[string][]string `json:"connections"`

	valid bool
}

type PortDef struct {
	Type   string             `json:"type"`
	Stream *PortDef           `json:"stream"`
	Map    map[string]PortDef `json:"map"`
	Any    string             `json:"any"`

	valid bool
}

// PUBLIC METHODS

func (d *InstanceDef) SetOperatorDef(operatorDef OperatorDef) error {
	if !operatorDef.Valid() {
		return errors.New("operator definition not validated")
	}
	d.operatorDef = operatorDef
	return nil
}

func (d *InstanceDef) OperatorDef() OperatorDef {
	return d.operatorDef
}

func (d InstanceDef) Valid() bool {
	return d.valid
}

func (d OperatorDef) Valid() bool {
	return d.valid
}

func (d *PortDef) Valid() bool {
	return d.valid
}

func (d *InstanceDef) Validate() error {
	if d.Name == "" {
		return fmt.Errorf(`instance name may not be empty`)
	}

	if strings.Contains(d.Name, " ") {
		return fmt.Errorf(`operator instance name may not contain spaces: "%s"`, d.Name)
	}

	if d.Operator == "" {
		return errors.New(`operator may not be empty`)
	}

	if strings.Contains(d.Operator, " ") {
		return fmt.Errorf(`operator may not contain spaces: "%s"`, d.Operator)
	}

	d.valid = true
	return nil
}

func (d *OperatorDef) Validate() error {
	if err := d.In.Validate(); err != nil {
		return err
	}

	if err := d.Out.Validate(); err != nil {
		return err
	}

	alreadyUsedInsNames := make(map[string]bool)
	for _, insDef := range d.Operators {
		if err := insDef.Validate(); err != nil {
			return err
		}

		if _, ok := alreadyUsedInsNames[insDef.Name]; ok {
			return fmt.Errorf(`colliding instance names within same parent operator: "%s"`, insDef.Name)
		}
		alreadyUsedInsNames[insDef.Name] = true
	}

	d.valid = true
	return nil
}

func (d *PortDef) Validate() error {
	if d.Type == "" {
		return errors.New("type must not be empty")
	}

	validTypes := []string{"any", "primitive", "number", "string", "boolean", "stream", "map"}
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

	if d.Type == "any" {
		if d.Any == "" {
			return errors.New("any identifier missing")
		}
	} else if d.Type == "stream" {
		if d.Stream == nil {
			return errors.New("stream missing")
		}
		return d.Stream.Validate()
	} else if d.Type == "map" {
		if len(d.Map) == 0 {
			return errors.New("map missing or empty")
		}
		for _, e := range d.Map {
			err := e.Validate()
			if err != nil {
				return err
			}
		}
	}

	d.valid = true
	return nil
}

func (d PortDef) SpecifyAnyPort(identifier string, def PortDef) (PortDef, error) {
	if d.Any == identifier {
		return def, nil
	}

	portDef := PortDef{}
	portDef.Type = d.Type

	if d.Type == "stream" {
		if portStr, err := d.Stream.SpecifyAnyPort(identifier, def); err == nil {
			portDef.Stream = &portStr
			return portDef, nil
		} else {
			return portDef, err
		}
	} else if d.Type == "map" {
		portDef.Map = make(map[string]PortDef)
		for k, e := range d.Map {
			var err error
			portDef.Map[k], err = e.SpecifyAnyPort(identifier, def)
			if err != nil {
				return portDef, err
			}
		}
	}

	return portDef, nil
}

func (d PortDef) FreeOfAnys() error {
	if d.Type == "any" || d.Any != "" {
		return errors.New("generic any not replaced: " + d.Any)
	}

	if d.Type == "stream" {
		return d.Stream.FreeOfAnys()
	} else if d.Type == "map" {
		for _, e := range d.Map {
			if err := e.FreeOfAnys(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d PortDef) Equals(p PortDef) bool {
	if d.Type != p.Type {
		return false
	}

	if d.Type == "map" {
		if len(d.Map) != len(p.Map) {
			return false
		}

		for k, e := range d.Map {
			pe, ok := p.Map[k]
			if !ok {
				return false
			}
			if !e.Equals(pe) {
				return false
			}
		}
	} else if d.Type == "stream" {
		if !d.Stream.Equals(*p.Stream) {
			return false
		}
	}

	return true
}
