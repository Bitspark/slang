package main

import (
	"fmt"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type testInstanceManager struct {
	list []*InstanceDef
}

func (m testInstanceManager) List() []InstanceDef {
	return funk.Map(funk.Filter(m.list, func(i *InstanceDef) bool {
		return i != nil
	}), func(i *InstanceDef) InstanceDef {
		return *i
	}).([]InstanceDef)
}

func (m *testInstanceManager) Start(instance uuid.UUID, mode T_DeploymentMode, def core.SlangFileDef) (InstanceDef, error) {
	ins := InstanceDef{instance, def.Main, Running, mode}
	m.list = append(m.list, &ins)
	return ins, nil
}
func (m testInstanceManager) Restart(instance uuid.UUID) error {
	return nil
}

func (m testInstanceManager) Stop(instance uuid.UUID) error {
	if funk.Find(m.List(), func(i *InstanceDef) bool { return i.Instance == instance }) != nil {
		return fmt.Errorf("unknown instance")
	}
	return nil
}

func (m testInstanceManager) Info(instance uuid.UUID) (InstanceDef, error) {
	if ins := funk.Find(m.List(), func(i InstanceDef) bool { return i.Instance == instance }); ins != nil {
		return ins.(InstanceDef), nil
	}
	return InstanceDef{uuid.Nil, uuid.Nil, "", ""}, fmt.Errorf("unknown instance")
}

func newTestDeployer() Deployer {
	return newDeployer(&testInstanceManager{})
}
