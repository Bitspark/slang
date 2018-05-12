package main

import (
	"log"
	"github.com/Bitspark/slang/pkg/daemon"
)

const PORT = 5149 // sla[n]g == 5149

func main() {
	log.Println("Starting slangd...")
	srv := daemon.New("localhost", PORT)
	srv.AddService("/operator", daemon.OperatorDefService)
	log.Printf("Listening on http://%s:%d/\n", srv.Host, srv.Port)
	log.Fatal(srv.Run())
}
