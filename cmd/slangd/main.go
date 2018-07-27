package main

import (
	"log"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/Bitspark/slang/pkg/daemon"
)

const PORT = 5149 // sla[n]g == 5149

type EnvironPaths struct {
	SLANG_PATH string
	SLANG_DIR  string
	SLANG_LIB  string
	SLANG_UI   string
}

func main() {
	log.Println("Starting slangd...")

	envPaths := initEnvironPaths()

	srv := daemon.New("localhost", PORT)

	envPaths.loadLocalComponents()
	envPaths.loadDaemonServices(srv)
	envPaths.startDaemonServer(srv)
}

func initEnvironPaths() (*EnvironPaths) {
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	slangPath := filepath.Join(currUser.HomeDir, "slang")

	e := &EnvironPaths{
		slangPath,
		daemon.EnsureEnvironVar("SLANG_DIR", filepath.Join(slangPath, "projects")),
		daemon.EnsureEnvironVar("SLANG_LIB", filepath.Join(slangPath, "lib")),
		daemon.EnsureEnvironVar("SLANG_UI", filepath.Join(slangPath, "ui")),
	}
	if _, err = daemon.EnsureDirExists(e.SLANG_DIR); err != nil {
		log.Fatal(err)
	}
	if _, err = daemon.EnsureDirExists(e.SLANG_LIB); err != nil {
		log.Fatal(err)
	}
	if _, err = daemon.EnsureDirExists(e.SLANG_UI); err != nil {
		log.Fatal(err)
	}
	return e
}

func (e *EnvironPaths) loadLocalComponents() {
	for repoName, dirPath := range map[string]string{"slang-lib": e.SLANG_LIB, "slang-ui": e.SLANG_UI} {
		dl := daemon.NewComponentLoader(repoName, dirPath)
		if dl.NewerVersionExists() {
			localVer := dl.GetLocalReleaseVersion()
			latestVer := dl.GetLatestReleaseVersion()
			if localVer != nil {
				log.Printf("Your local %v has version %v but latest is %v.", repoName, localVer.String(), latestVer.String())
			}
			log.Printf("Downloading %v latest version (%v).", repoName, latestVer.String())

			if err := dl.Load(); err != nil {
				log.Fatal(err)
			}
			log.Printf("Done.")
		} else {
			localVer := dl.GetLocalReleaseVersion()
			log.Printf("Your local %v is up-to-date (%v).", repoName, localVer.String())

		}

	}
}

func (e *EnvironPaths) loadDaemonServices(srv *daemon.DaemonServer) {
	srv.AddRedirect("/", "/app")
	srv.AddStaticServer("/app", http.Dir(e.SLANG_UI))
	srv.AddService("/operator", daemon.DefinitionService)
	srv.AddService("/run", daemon.RunnerService)
}

func (e *EnvironPaths) startDaemonServer(srv *daemon.DaemonServer) {
	log.Printf("\n\n\tListening on http://%s:%d/\n\n", srv.Host, srv.Port)
	log.Fatal(srv.Run())
}
