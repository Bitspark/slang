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
	Generics   map[string]*PortDef    `json:"generics"`

	valid          bool
	operatorDef    OperatorDef
	operatorDefSet bool
}

type OperatorDef struct {
	In          PortDef             `json:"in"`
	Out         PortDef             `json:"out"`
	Operators   []*InstanceDef      `json:"operators"`
	Connections map[string][]string `json:"connections"`

	valid bool
}

type PortDef struct {
	// Type is one of "primitive", "number", "string", "boolean", "stream", "map", "generic"
	Type    string              `json:"type"`
	Stream  *PortDef            `json:"stream"`
	Map     map[string]*PortDef `json:"map"`
	Generic string              `json:"generic"`

	valid bool
}

// INSTANCE DEFINITION

func (d InstanceDef) Valid() bool {
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

func (d InstanceDef) OperatorDef() OperatorDef {
	return d.operatorDef
}

func (d InstanceDef) HasOperatorDef() bool {
	return d.operatorDefSet
}

func (d *InstanceDef) SetOperatorDef(operatorDef OperatorDef) error {
	d.operatorDef = operatorDef
	d.operatorDefSet = true
	return nil
}

// OPERATOR DEFINITION

func (d OperatorDef) Valid() bool {
	return d.valid
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

// SpecifyGenericPorts specifies generic types in the operator definition. It does not touch referenced values
// such as *PortDef but replaces them with a reference on a copy.
func (d *OperatorDef) SpecifyGenericPorts(generics map[string]*PortDef) error {
	if err := d.In.SpecifyGenericPorts(generics); err != nil {
		return err
	}
	if err := d.Out.SpecifyGenericPorts(generics); err != nil {
		return err
	}
	for _, op := range d.Operators {
		for _, gp := range op.Generics {
			if err := gp.SpecifyGenericPorts(generics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d OperatorDef) GenericsSpecified() error {
	if err := d.In.GenericsSpecified(); err != nil {
		return err
	}
	if err := d.Out.GenericsSpecified(); err != nil {
		return err
	}
	for _, op := range d.Operators {
		for _, gp := range op.Generics {
			if err := gp.GenericsSpecified(); err != nil {
				return err
			}
		}
	}
	return nil
}

// PORT DEFINITION

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
			if !e.Equals(*pe) {
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

func (d *PortDef) Valid() bool {
	return d.valid
}

func (d *PortDef) Validate() error {
	if d.Type == "" {
		return errors.New("type must not be empty")
	}

	validTypes := []string{"generic", "primitive", "number", "string", "boolean", "stream", "map"}
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

	if d.Type == "generic" {
		if d.Generic == "" {
			return errors.New("generic identifier missing")
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
			if e == nil {
				return errors.New("map entry must not be null")
			}
			err := e.Validate()
			if err != nil {
				return err
			}
		}
	}

	d.valid = true
	return nil
}

func (d PortDef) Copy() PortDef {
	cpy := PortDef{Type: d.Type, Generic: d.Generic}

	if d.Stream != nil {
		strCpy := d.Stream.Copy()
		cpy.Stream = &strCpy
	}
	if d.Map != nil {
		cpy.Map = make(map[string]*PortDef)
		for k, e := range d.Map {
			subCpy := e.Copy()
			cpy.Map[k] = &subCpy
		}
	}

	return cpy
}

// SpecifyGenericPorts specifies generic types in the port definition. It does not touch referenced values
// such as *PortDef but replaces them with a reference on a copy, which is very important to prevent unintended side
// effects.
func (d *PortDef) SpecifyGenericPorts(generics map[string]*PortDef) error {
	for identifier, pd := range generics {
		if d.Generic == identifier {
			// Replace with copy!
			*d = pd.Copy()
			return nil
		}

		if d.Type == "stream" {
			strCpy := d.Stream.Copy()
			// Replace with copy!
			d.Stream = &strCpy
			return strCpy.SpecifyGenericPorts(generics)
		} else if d.Type == "map" {
			for k, e := range d.Map {
				eCpy := e.Copy()
				if err := eCpy.SpecifyGenericPorts(generics); err != nil {
					return err
				}
				// Replace with copy!
				d.Map[k] = &eCpy
			}
		}
	}
	return nil
}

func (d PortDef) GenericsSpecified() error {
	if d.Type == "generic" || d.Generic != "" {
		return errors.New("generic not replaced: " + d.Generic)
	}

	if d.Type == "stream" {
		return d.Stream.GenericsSpecified()
	} else if d.Type == "map" {
		for _, e := range d.Map {
			if err := e.GenericsSpecified(); err != nil {
				return err
			}
		}
	}

	return nil
}
