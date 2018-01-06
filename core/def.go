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
	In         *PortDef               `json:"in"`
	Out        *PortDef               `json:"out"`
	valid      bool
}

type OperatorDef struct {
	In          *PortDef            `json:"in"`
	Out         *PortDef            `json:"out"`
	Operators   []InstanceDef       `json:"operators"`
	Connections map[string][]string `json:"connections"`
	valid       bool
}

type PortDef struct {
	Type   string             `json:"type"`
	Stream *PortDef           `json:"stream"`
	Map    map[string]PortDef `json:"map"`
	Any    string             `json:"any"`
	valid  bool
}

// PUBLIC METHODS

func (d InstanceDef) Valid() bool {
	return d.valid
}

func (d OperatorDef) Valid() bool {
	return d.valid
}

func (d *PortDef) Valid() bool {
	return d.valid
}

func (d *PortDef) Primitive() bool {
	return d.Type == "number" || d.Type == "string" || d.Type == "boolean" || d.Type == "primitive"
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

	if d.In != nil {
		if err := d.In.Validate(); err != nil {
			return err
		}
	}

	if d.Out != nil {
		if err := d.Out.Validate(); err != nil {
			return err
		}
	}

	d.valid = true
	return nil
}

func (d *OperatorDef) Validate() error {
	if d.In == nil || d.Out == nil {
		return errors.New(`ports must be defined`)
	}

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
		if len(d.Any) == 0 {
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

func (d PortDef) Equals(p PortDef) error {
	if d.Type != p.Type {
		return errors.New(fmt.Sprintf("different types: %s != %s", d.Type, p.Type))
	}

	if d.Type == "map" {
		if len(d.Map) != len(p.Map) {
			return errors.New(fmt.Sprintf("different map lengths: %d != %d", len(d.Map), len(p.Map)))
		}

		for k, e := range d.Map {
			pe, ok := p.Map[k]
			if !ok {
				return errors.New(fmt.Sprintf("key not present: %s", k))
			}
			if err := e.Equals(pe); err != nil {
				return err
			}
		}
	} else if d.Type == "stream" {
		if err := d.Stream.Equals(*p.Stream); err != nil {
			return err
		}
	}

	return nil
}
