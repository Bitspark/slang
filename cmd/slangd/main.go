package main

import (
	"log"
	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/api"
)

const PORT = 5149 // sla[n]g == 5149

func main() {
	// call NewEnviron to check, whether Environ can be correctly loaded
	api.NewEnviron()
	log.Println("Starting slangd...")
	srv := daemon.New("localhost", PORT)
	srv.AddService("/operator", daemon.DefinitionService)
	srv.AddService("/run", daemon.RunnerService)
	log.Printf("Listening on http://%s:%d/\n", srv.Host, srv.Port)
	log.Fatal(srv.Run())
}
