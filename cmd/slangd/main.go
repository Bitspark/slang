package main

import (
	"log"
	"github.com/Bitspark/slang/pkg/daemon"
	"os/user"
	"path/filepath"
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

	loadLocalComponents(envPaths)
	loadDaemonServices(srv)
	startDaemonServer(srv)
}

func initEnvironPaths() (*EnvironPaths) {
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	slangPath := filepath.Join(currUser.HomeDir, "slang")
	return &EnvironPaths{
		slangPath,
		daemon.EnsureEnvironVar("SLANG_DIR", filepath.Join(slangPath, "projects")),
		daemon.EnsureEnvironVar("SLANG_LIB", filepath.Join(slangPath, "lib")),
		daemon.EnsureEnvironVar("SLANG_UI", filepath.Join(slangPath, "ui")),
	}
}

func loadLocalComponents(envPaths *EnvironPaths) {
	for repoName, dirPath := range map[string]string{"slang-lib": envPaths.SLANG_LIB, "slang-ui": envPaths.SLANG_UI} {
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

func loadDaemonServices(srv *daemon.DaemonServer) {
	srv.AddService("/operator", daemon.DefinitionService)
	srv.AddService("/run", daemon.RunnerService)
}

func startDaemonServer(srv *daemon.DaemonServer) {
	log.Printf("\n\n\tListening on http://%s:%d/\n\n", srv.Host, srv.Port)
	log.Fatal(srv.Run())
}
