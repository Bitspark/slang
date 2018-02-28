package main

import (
	"log"
	"github.com/Bitspark/slang/pkg/server"
)

const PORT = 5149 // sla[n]g == 5149

func main() {
	log.Println("Starting slangd...")
	srv := server.New(PORT)
	//srv.AddEndpoint("/builtin", &server.ListBuiltinNames{})
	log.Printf("Listening on http://localhost:%v/\n", srv.Port)
	log.Fatal(srv.Run())
}
