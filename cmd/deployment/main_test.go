package main

import (
	"fmt"
	"testing"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func (m testInstanceManager) Start(instance uuid.UUID, mode DeploymentMode, def core.SlangFileDef) (InstanceDef, error) {
	ins := InstanceDef{uuid.New(), def.Main, Running, mode}
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
	if ins := funk.Find(m.List(), func(i *InstanceDef) bool { return i.Instance == instance }).(*InstanceDef); ins != nil {
		return *ins, nil
	}
	return InstanceDef{uuid.Nil, uuid.Nil, "", ""}, fmt.Errorf("unknown instance")
}

func newTestDeploymentHandler() DeploymentHandler {
	return newDeploymentHandler(&testInstanceManager{})
}

func TestDeploymentHandler_Deploy(t *testing.T) {
	dh := newTestDeploymentHandler()

	assert.Empty(t, dh.List())
	assert.NotEmpty(t, dh.List())
}
