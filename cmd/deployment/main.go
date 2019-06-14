package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type T_RuntimeState string

const (
	Spawning T_RuntimeState = "spawning"
	Running  T_RuntimeState = "running"
	Stopped  T_RuntimeState = "stopped"
)

type T_DeploymentMode string

const (
	Rest    T_DeploymentMode = "rest"
	Form    T_DeploymentMode = "form"
	Process T_DeploymentMode = "proc"
)

type InstanceDef struct {
	Instance uuid.UUID
	Operator uuid.UUID
	State    T_RuntimeState
	Mode     T_DeploymentMode
}

type InstanceManager interface {
	List() []InstanceDef
	Start(instance uuid.UUID, mode T_DeploymentMode, def core.SlangFileDef) (InstanceDef, error)
	Restart(instance uuid.UUID) error
	Stop(instance uuid.UUID) error
	Info(instance uuid.UUID) (InstanceDef, error)
}

type Deployer interface {
	List() []InstanceDef
	Deploy(def core.SlangFileDef, mode T_DeploymentMode) (InstanceDef, error)
	Restart(instance uuid.UUID) error
	Shutdown(instance uuid.UUID) error
	Get(instance uuid.UUID) (InstanceDef, error)
}

type deployerImpl struct {
	pm InstanceManager
}

func newDeployer(pm InstanceManager) Deployer {
	return &deployerImpl{pm}
}

func (d deployerImpl) List() []InstanceDef {
	return d.pm.List()
}

func (d deployerImpl) Deploy(def core.SlangFileDef, mode T_DeploymentMode) (InstanceDef, error) {
	instance := uuid.New()
	return d.pm.Start(instance, mode, def)

}

func (d deployerImpl) Shutdown(instance uuid.UUID) error {
	return d.pm.Stop(instance)
}

func (d deployerImpl) Restart(instance uuid.UUID) error {
	return d.pm.Restart(instance)
}

func (d deployerImpl) Get(instance uuid.UUID) (InstanceDef, error) {
	return d.pm.Info(instance)
}

func main() {
	//portNo := flag.Int("port", 80, "Choose server port number")
	flag.Parse()

	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNo), newRouter()))
}

func addContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getDeployer(r *http.Request) Deployer {
	return r.Context().Value("deployer").(Deployer)
}

func newRouter(deployer Deployer) http.Handler {
	r := mux.NewRouter()

	registerHandler(r)

	ctx := context.WithValue(context.Background(), "deployer", deployer)

	return addContext(ctx, r)
}

func registerHandler(r *mux.Router) {
	r.PathPrefix("/api/v1/modes").Methods("GET").HandlerFunc(listDeploymentModes)
	r.PathPrefix("/api/v1/instances").Methods("POST").HandlerFunc(deployInstance)
}
