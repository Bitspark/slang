package main

import (
	"flag"
	"net/http"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var portNo int

type RuntimeState string

const (
	Spawning RuntimeState = "spawning"
	Running  RuntimeState = "running"
	Stopped  RuntimeState = "stopped"
)

type DeploymentMode string

const (
	Rest    DeploymentMode = "rest"
	Form    DeploymentMode = "form"
	Process DeploymentMode = "proc"
)

type InstanceDef struct {
	Instance uuid.UUID
	Operator uuid.UUID
	State    RuntimeState
	Mode     DeploymentMode
}

type InstanceManager interface {
	List() []InstanceDef
	Start(instance uuid.UUID, mode DeploymentMode, def core.SlangFileDef) (InstanceDef, error)
	Restart(instance uuid.UUID) error
	Stop(instance uuid.UUID) error
	Info(instance uuid.UUID) (InstanceDef, error)
}

type DeploymentHandler interface {
	List() []InstanceDef
	Deploy(def core.SlangFileDef, mode DeploymentMode) (InstanceDef, error)
	Restart(instance uuid.UUID) error
	Shutdown(instance uuid.UUID) error
	Get(instance uuid.UUID) (InstanceDef, error)
}

type deploymentHandler struct {
	pm InstanceManager
}

func newDeploymentHandler(pm InstanceManager) DeploymentHandler {
	return &deploymentHandler{pm}
}

func (d deploymentHandler) List() []InstanceDef {
	return d.pm.List()
}

func (d deploymentHandler) Deploy(def core.SlangFileDef, mode DeploymentMode) (InstanceDef, error) {
	instance := uuid.New()
	return d.pm.Start(instance, mode, def)

}

func (d deploymentHandler) Shutdown(instance uuid.UUID) error {
	return d.pm.Stop(instance)
}

func (d deploymentHandler) Restart(instance uuid.UUID) error {
	return d.pm.Restart(instance)
}

func (d deploymentHandler) Get(instance uuid.UUID) (InstanceDef, error) {
	return d.pm.Info(instance)
}

func main() {
	flag.IntVar(&portNo, "port", 80, "Choose server port number")
	flag.Parse()

	r := mux.NewRouter()
	registerHandlerV1(r.PathPrefix("/api/v1").Subrouter())
}

func registerHandlerV1(v1 *mux.Router) {
	v1Instances := v1.PathPrefix("/instances").Subrouter()
	v1Instances.Methods("POST").HandlerFunc(func(respose http.ResponseWriter, request *http.Request) {
	})
}
