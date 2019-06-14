package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func responseWithError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		panic(err)
	}
}

func responseWithOk(w http.ResponseWriter, m interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		panic(err)
	}
}

func getBodyBufferSafe(r *http.Request) []byte {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 10485760)) // limit to 10 MB
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	return body
}

type DeploymentMode struct {
	Mode        T_DeploymentMode
	Title       string
	Description string
}

type DeploymentModes []DeploymentMode

func listDeploymentModes(w http.ResponseWriter, r *http.Request) {
	m := DeploymentModes{
		{Mode: Rest, Title: "REST", Description: "REST"},
		{Mode: Form, Title: "FORM", Description: "FORM"},
		{Mode: Process, Title: "PROCESS", Description: "PROCESS"},
	}

	responseWithOk(w, m)
}

type DeployInstruction struct {
	SlangFile core.SlangFileDef
	Mode      T_DeploymentMode
}

func deployInstance(w http.ResponseWriter, r *http.Request) {
	body := getBodyBufferSafe(r)

	var instruction DeployInstruction
	if err := json.Unmarshal(body, &instruction); err != nil {
		responseWithError(w, err, http.StatusBadRequest)
		return
	}

	deployer := getDeployer(r)

	instDef, err := deployer.Deploy(instruction.SlangFile, instruction.Mode)

	if err != nil {
		responseWithError(w, err, http.StatusBadRequest)
		return
	}

	responseWithOk(w, instDef)
}

func getInstances(w http.ResponseWriter, r *http.Request) {
	deployer := getDeployer(r)

	vars := mux.Vars(r)
	instance, err := uuid.Parse(vars["instance"])

	if err != nil {
		responseWithError(w, err, http.StatusBadRequest)
	}

	instDef, err := deployer.Get(instance)

	if err != nil {
		responseWithError(w, err, http.StatusBadRequest)
		return
	}

	responseWithOk(w, instDef)
}

func listInstances(w http.ResponseWriter, r *http.Request) {
	deployer := getDeployer(r)

	instances := deployer.List()

	responseWithOk(w, instances)
}
