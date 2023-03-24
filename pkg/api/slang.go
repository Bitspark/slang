package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
)

var PROPERTY_PLACERHOLDER_REGEXP = regexp.MustCompile(`(\$\w+)`)

// todo should be SlangBundle method
func BuildOperator(bundle *core.SlangBundle) (*core.Operator, error) {
	if !bundle.Valid() {
		if err := bundle.Validate(); err != nil {
			return nil, err
		}
	}

	stor := newSlangBundleStorage(funk.Values(bundle.Blueprints).([]core.Blueprint))

	return BuildAndCompile(bundle.Main, bundle.Args.Generics, bundle.Args.Properties, *stor)
}

func gatherDependencies(def *core.Blueprint, bundle *core.SlangBundle, store *storage.Storage) error {
	bundle.Blueprints[def.Id] = *def
	for _, dep := range def.InstanceDefs {
		id := dep.Operator
		if _, ok := bundle.Blueprints[id]; !ok {
			depDef, err := store.Load(id)
			if err != nil {
				return err
			}
			bundle.Blueprints[id] = *depDef
			err = gatherDependencies(depDef, bundle, store)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateBundle(bp *core.Blueprint, st *storage.Storage) (*core.SlangBundle, error) {
	bundle := &core.SlangBundle{
		Main:       bp.Id,
		Blueprints: make(map[uuid.UUID]core.Blueprint),
	}

	if err := gatherDependencies(bp, bundle, st); err != nil {
		return bundle, err
	}

	return bundle, bundle.Validate()
}

type slangBundleLoader struct {
	blueprintById map[uuid.UUID]core.Blueprint
}

func newSlangBundleStorage(blueprints []core.Blueprint) *storage.Storage {
	m := make(map[uuid.UUID]core.Blueprint)

	for _, bp := range blueprints {
		m[bp.Id] = bp
	}

	return storage.NewStorage().AddBackend(&slangBundleLoader{m})
}

func (l *slangBundleLoader) Has(opId uuid.UUID) bool {
	_, ok := l.blueprintById[opId]
	return ok
}

func (l *slangBundleLoader) List() ([]uuid.UUID, error) {
	var uuidList []uuid.UUID

	for _, idOrName := range funk.Keys(l.blueprintById).([]string) {
		if id, err := uuid.Parse(idOrName); err == nil {
			uuidList = append(uuidList, id)
		}
	}

	return uuidList, nil
}

func (l *slangBundleLoader) Load(opId uuid.UUID) (*core.Blueprint, error) {
	if blueprint, ok := l.blueprintById[opId]; ok {
		return &blueprint, nil
	}
	return nil, fmt.Errorf("unknown operator")
}

func CreateAndConnectOperator(insName string, def core.Blueprint, ordered bool) (*core.Operator, error) {
	// Create new non-builtin operator
	o, err := core.NewOperator(insName, nil, nil, nil, nil, def)
	if err != nil {
		return nil, err
	}

	// Recursively create all child operators from top to bottom
	for _, childOpInsDef := range def.InstanceDefs {
		if builtinOp, err := elem.MakeOperator(*childOpInsDef); err == nil {
			// Builtin operator has been found
			builtinOp.SetParent(o)
			continue
		} else if elem.IsRegistered(childOpInsDef.Operator) {
			// Builtin operator with that name exists, but still could not create it, so an error must have occurred
			return nil, err
		}

		oc, err := CreateAndConnectOperator(childOpInsDef.Name, childOpInsDef.Blueprint, ordered)
		if err != nil {
			return nil, err
		}

		oc.SetParent(o)
	}

	// Parse all connections before starting to connect
	parsedConns := make(map[*core.Port][]*core.Port)
	for srcConnDef, dstConnDefs := range def.Connections {
		if pSrc, err := core.ParsePortReference(srcConnDef, o); err == nil {
			parsedConns[pSrc] = nil
			for _, dstConnDef := range dstConnDefs {
				if pDst, err := core.ParsePortReference(dstConnDef, o); err == nil {
					parsedConns[pSrc] = append(parsedConns[pSrc], pDst)
				} else {
					return nil, fmt.Errorf("%s: %s", err.Error(), dstConnDef)
				}
			}
		} else {
			return nil, fmt.Errorf("%s: %s", err.Error(), srcConnDef)
		}
	}

	if err := connectDestinations(o, parsedConns, ordered); err != nil {
		return nil, err
	}

	return o, nil
}

// connectDestinations connects operators following from the in port to the out port
func connectDestinations(o *core.Operator, conns map[*core.Port][]*core.Port, ordered bool) error {
	var ops []*core.Operator
	for pSrc, pDsts := range conns {
		if pSrc.Operator() != o {
			continue
		}
		// Start with operator o
		for _, pDst := range pDsts {
			if err := pSrc.Connect(pDst); err != nil {
				return fmt.Errorf("%s -> %s: %s", pSrc.Name(), pDst.Name(), err)
			}
			ops = append(ops, pDst.Operator())
		}
		// Set the destinations nil so that we do not end in an infinite recursion
		conns[pSrc] = nil
	}

	var contdOps []*core.Operator
	if ordered {
		// Filter for ops that have all in ports connected
		for _, op := range ops {
			connected := true
			for _, pDsts := range conns {
				for _, pDst := range pDsts {
					if op == pDst.Operator() && pDst.Delegate() == nil {
						connected = false
						goto end
					}
				}
			}
		end:
			if connected {
				contdOps = append(contdOps, op)
			}
		}
	} else {
		contdOps = ops
	}

	// Continue with ops that are completely connected
	for _, op := range contdOps {
		if err := connectDestinations(op, conns, ordered); err != nil {
			return err
		}
	}
	return nil
}

func BuildAndCompile(bpid uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*core.Operator, error) {
	if op, err := Build(bpid, gens, props, st); err == nil {
		return Compile(op)
	} else {
		return op, err
	}
}

func Build(bpid uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*core.Operator, error) {
	if !elem.Initalized {
		return nil, fmt.Errorf("call elem.Init() before api.Build() or api.BuildAndCompile()")
	}

	// Recursively replace generics by their actual types and propagate properties
	// TODO SpecifyOperator should instantiate and return an Operator
	blueprint, err := st.Load(bpid)

	if err != nil {
		return nil, err
	}

	err = specifyOperator(blueprint, gens, props, st, []uuid.UUID{})
	if err != nil {
		return nil, err
	}

	// Create and connect the operator
	op, err := CreateAndConnectOperator("", *blueprint, false)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func Compile(op *core.Operator) (*core.Operator, error) {
	// Compile
	op.Compile()

	// Connect
	flatDef, err := op.Define()
	if err != nil {
		return nil, err
	}

	// Create and connect the flat operator
	flatOp, err := CreateAndConnectOperator("", flatDef, true)
	if err != nil {
		return nil, err
	}

	// Check if all in ports are connected
	err = flatOp.CorrectlyCompiled()
	if err != nil {
		return nil, err
	}

	return flatOp, nil
}

func completeProperties(blueprint *core.Blueprint, givenProps core.Properties) (core.Properties, error) {
	propDefs := blueprint.PropertyDefs
	completedProps := make(core.Properties)

	for propKey, propDef := range propDefs {
		if prop, err := givenProps.Get(propKey, propDef); err == nil {
			completedProps[propKey] = prop
		} else {
			return nil, err
		}
	}

	return completedProps, nil
}

func specifyOperator(blueprint *core.Blueprint, gens core.Generics, props core.Properties, st storage.Storage, dependencyChain []uuid.UUID) error {
	var err error

	if props, err = completeProperties(blueprint, props); err != nil {
		return err
	}

	if err := blueprint.SpecifyOperator(gens, props); err != nil {
		return err
	}
	dependencyChain = append(dependencyChain, blueprint.Id)

	for _, childInsDef := range blueprint.InstanceDefs {

		// Load Blueprint for childInsDef
		if childInsDef.Blueprint.Id == uuid.Nil {
			childOpId := childInsDef.Operator
			if childBlueprint, err := st.Load(childOpId); err == nil {
				childInsDef.Blueprint = *childBlueprint
			} else {
				return err
			}
		}

		if funk.Contains(dependencyChain, childInsDef.Operator) {
			return fmt.Errorf("recursion in %s", blueprint.Id)
		}

		// Propagate property values to child operators
		for prop, propVal := range childInsDef.Properties {
			updated, newPropVal, err := interpolatePropVal(propVal, props)

			if err != nil {
				return err
			}

			if !updated {
				continue
			}

			childInsDef.Properties[prop] = newPropVal
		}

		for _, gen := range childInsDef.Generics {
			gen.SpecifyGenerics(gens)
		}

		err := specifyOperator(&childInsDef.Blueprint, childInsDef.Generics, childInsDef.Properties, st, dependencyChain)

		if err != nil {
			return err
		}
	}

	blueprint.PropertyDefs = nil

	return nil
}

func interpolatePropVal(propVal interface{}, props core.Properties) (bool, interface{}, error) {
	propStr, isString := propVal.(string)
	if isString {

		for _, propKey := range PROPERTY_PLACERHOLDER_REGEXP.FindAllString(propStr, -1) {
			if val, ok := props[propKey[1:]]; ok {
				val, ok := val.(string)

				if !ok {
					return false, propVal, fmt.Errorf("cannot interpolate \"%s\" with value \"%v\"", propKey, val)
				}

				propStr = strings.Replace(propStr, propKey, val, -1)
			} else {
				return false, propVal, fmt.Errorf("unknown property \"%s\"", propKey)
			}
		}

		return true, propStr, nil
	}

	propMapVal, isMap := propVal.(map[string]interface{})
	if isMap {
		anyUpdated := false
		for k, v := range propMapVal {
			updated, new, err := interpolatePropVal(v, props)

			if err != nil {
				return false, propVal, err
			}

			if !updated {
				continue
			}

			if !anyUpdated {
				anyUpdated = true
			}

			propMapVal[k] = new
		}

		return anyUpdated, propMapVal, nil
	}

	propArrayVal, isArray := propVal.([]interface{})
	if isArray {
		anyUpdated := false
		for i, v := range propArrayVal {
			updated, new, err := interpolatePropVal(v, props)

			if err != nil {
				return false, propVal, err
			}

			if !updated {
				continue
			}

			if !anyUpdated {
				anyUpdated = true
			}

			propArrayVal[i] = new
		}

		return anyUpdated, propArrayVal, nil
	}

	return false, propVal, nil
}
