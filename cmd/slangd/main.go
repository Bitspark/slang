package main

import (
	"log"
	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"os/user"
	"path/filepath"
)

const PORT = 5149 // sla[n]g == 5149

func getEnviron() *api.Environ {

	currUser, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	slangPath := filepath.Join(currUser.HomeDir, "slang")
	core.EnsureEnvironVar("SLANG_DIR", filepath.Join(slangPath, "projects"))
	core.EnsureEnvironVar("SLANG_LIB", filepath.Join(slangPath, "lib"))
	core.EnsureEnvironVar("SLANG_UI", filepath.Join(slangPath, "ui"))

	return api.NewEnviron()
}

func main() {
	log.Println("Starting slangd...")
	env := getEnviron()
	srv := daemon.New(env, "localhost", PORT)

	srv.AddService("/operator", daemon.DefinitionService)
	srv.AddService("/run", daemon.RunnerService)

	log.Printf("Listening on http://%s:%d/\n", srv.Host, srv.Port)
	log.Fatal(srv.Run())
}
