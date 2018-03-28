package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Bitspark/slang/pkg/utils"
	"strings"
)

type InstanceDefList []*InstanceDef
type TypeDefMap map[string]TypeDef
type Properties utils.MapStr
type Generics map[string]*TypeDef

type InstanceDef struct {
	Operator   string     `json:"operator" yaml:"operator"`
	Name       string     `json:"name" yaml:"name"`
	Properties Properties `json:"properties" yaml:"properties"`
	Generics   Generics   `json:"generics" yaml:"generics"`

	valid       bool
	operatorDef OperatorDef
}

type OperatorDef struct {
	ServiceDefs  map[string]*ServiceDef  `json:"services" yaml:"services"`
	DelegateDefs map[string]*DelegateDef `json:"delegates" yaml:"delegates"`
	InstanceDefs InstanceDefList         `json:"operators" yaml:"operators"`
	PropertyDefs TypeDefMap              `json:"properties" yaml:"properties"`
	Connections  map[string][]string     `json:"connections" yaml:"connections"`

	valid bool
}

type DelegateDef struct {
	In  TypeDef `json:"in" yaml:"in"`
	Out TypeDef `json:"out" yaml:"out"`

	valid bool
}

type ServiceDef struct {
	In  TypeDef `json:"in" yaml:"in"`
	Out TypeDef `json:"out" yaml:"out"`

	valid bool
}

type TypeDef struct {
	// Type is one of "primitive", "number", "string", "boolean", "stream", "map", "generic"
	Type    string              `json:"type" yaml:"type"`
	Stream  *TypeDef            `json:"stream" yaml:"stream"`
	Map     map[string]*TypeDef `json:"map" yaml:"map"`
	Generic string              `json:"generic" yaml:"generic"`

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

func (d *InstanceDef) SetOperatorDef(operatorDef OperatorDef) error {
	d.operatorDef = operatorDef
	return nil
}

// OPERATOR DEFINITION

func (d OperatorDef) Valid() bool {
	return d.valid
}

func (d *OperatorDef) Validate() error {
	for _, srv := range d.ServiceDefs {
		if err := srv.Validate(); err != nil {
			return err
		}
	}

	for _, del := range d.DelegateDefs {
		if err := del.Validate(); err != nil {
			return err
		}
	}

	alreadyUsedInsNames := make(map[string]bool)
	for _, insDef := range d.InstanceDefs {
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

// SpecifyGenerics replaces generic types in the operator definition with the types given in the generics map.
// The values of the map are the according identifiers. It does not touch referenced values such as *PortDef but
// replaces them with a reference on a copy.
func (d *OperatorDef) SpecifyGenericPorts(generics map[string]*TypeDef) error {
	srvs := make(map[string]*ServiceDef)
	for srvName := range d.ServiceDefs {
		srv := d.ServiceDefs[srvName].Copy()
		if err := srv.In.SpecifyGenerics(generics); err != nil {
			return err
		}
		if err := srv.Out.SpecifyGenerics(generics); err != nil {
			return err
		}
		srvs[srvName] = &srv
	}
	d.ServiceDefs = srvs

	dels := make(map[string]*DelegateDef)
	for delName := range d.DelegateDefs {
		del := d.DelegateDefs[delName].Copy()
		if err := del.In.SpecifyGenerics(generics); err != nil {
			return err
		}
		if err := del.Out.SpecifyGenerics(generics); err != nil {
			return err
		}
		dels[delName] = &del
	}
	d.DelegateDefs = dels
	for _, op := range d.InstanceDefs {
		for _, gp := range op.Generics {
			if err := gp.SpecifyGenerics(generics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d OperatorDef) GenericsSpecified() error {
	for _, srv := range d.ServiceDefs {
		if err := srv.In.GenericsSpecified(); err != nil {
			return err
		}
		if err := srv.Out.GenericsSpecified(); err != nil {
			return err
		}
	}
	for _, del := range d.DelegateDefs {
		if err := del.In.GenericsSpecified(); err != nil {
			return err
		}
		if err := del.Out.GenericsSpecified(); err != nil {
			return err
		}
	}
	for _, op := range d.InstanceDefs {
		for _, gp := range op.Generics {
			if err := gp.GenericsSpecified(); err != nil {
				return err
			}
		}
	}
	return nil
}

// SERVICE DEFINITION

func (d *ServiceDef) Valid() bool {
	return d.valid
}

func (d *ServiceDef) Validate() error {
	if err := d.In.Validate(); err != nil {
		return err
	}

	if err := d.Out.Validate(); err != nil {
		return err
	}

	d.valid = true
	return nil
}

func (d ServiceDef) Copy() ServiceDef {
	cpy := ServiceDef{}

	cpy.In = d.In.Copy()
	cpy.Out = d.Out.Copy()

	return cpy
}

// DELEGATE DEFINITION

func (d *DelegateDef) Valid() bool {
	return d.valid
}

func (d *DelegateDef) Validate() error {
	if err := d.In.Validate(); err != nil {
		return err
	}

	if err := d.Out.Validate(); err != nil {
		return err
	}

	d.valid = true
	return nil
}

func (d DelegateDef) Copy() DelegateDef {
	cpy := DelegateDef{}

	cpy.In = d.In.Copy()
	cpy.Out = d.Out.Copy()

	return cpy
}

// TYPE DEFINITION

func (d TypeDef) Equals(p TypeDef) bool {
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

func (d *TypeDef) Valid() bool {
	return d.valid
}

func (d *TypeDef) Validate() error {
	if d.Type == "" {
		return errors.New("type must not be empty")
	}

	validTypes := []string{"generic", "primitive", "trigger", "number", "string", "binary", "boolean", "stream", "map"}
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

func (d TypeDef) Copy() TypeDef {
	cpy := TypeDef{Type: d.Type, Generic: d.Generic}

	if d.Stream != nil {
		strCpy := d.Stream.Copy()
		cpy.Stream = &strCpy
	}
	if d.Map != nil {
		cpy.Map = make(map[string]*TypeDef)
		for k, e := range d.Map {
			subCpy := e.Copy()
			cpy.Map[k] = &subCpy
		}
	}

	return cpy
}

// SpecifyGenerics replaces generic types in the port definition with the types given in the generics map.
// The values of the map are the according identifiers. It does not touch referenced values such as *PortDef but
// replaces them with a reference on a copy, which is very important to prevent unintended side effects.
func (d *TypeDef) SpecifyGenerics(generics map[string]*TypeDef) error {
	for identifier, pd := range generics {
		if d.Generic == identifier {
			*d = pd.Copy()
			return nil
		}

		if d.Type == "stream" {
			strCpy := d.Stream.Copy()
			d.Stream = &strCpy
			return strCpy.SpecifyGenerics(generics)
		} else if d.Type == "map" {
			mapCpy := make(map[string]*TypeDef)
			for k, e := range d.Map {
				eCpy := e.Copy()
				if err := eCpy.SpecifyGenerics(generics); err != nil {
					return err
				}
				mapCpy[k] = &eCpy
			}
			d.Map = mapCpy
		}
	}
	return nil
}

func (d TypeDef) GenericsSpecified() error {
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

func (d TypeDef) VerifyData(data interface{}) error {
	return nil
}

// TYPE DEF MAP

func (t TypeDefMap) VerifyData(data interface{}) error {
	return nil
}

// OPERATOR LIST MARSHALLING

func (ol *InstanceDefList) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var im map[string]*InstanceDef
	if err := unmarshal(&im); err != nil {
		return err
	}

	instances := make([]*InstanceDef, len(im))
	i := 0
	for name, inst := range im {
		inst.Name = name
		instances[i] = inst
		i++
	}

	*ol = instances
	return nil
}

func (ol *InstanceDefList) UnmarshalJSON(data []byte) error {
	var im map[string]*InstanceDef
	if err := json.Unmarshal(data, &im); err != nil {
		return err
	}

	instances := make([]*InstanceDef, len(im))
	i := 0
	for name, inst := range im {
		inst.Name = name
		instances[i] = inst
		i++
	}

	*ol = instances
	return nil
}
