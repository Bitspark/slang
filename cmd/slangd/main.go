package main

import (
	"fmt"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"time"

	"github.com/Bitspark/browser"
	"github.com/Bitspark/slang/pkg/daemon"
	"strconv"
)

const PORT = 5149 // sla[n]g == 5149

// will be set during build process
var (
	Version   string
	BuildTime string
)

type EnvironPaths struct {
	SLANG_PATH string
	SLANG_DIR  string
	SLANG_LIB  string
	SLANG_UI   string
}

func main() {
	buildTime, _ := strconv.ParseInt(BuildTime, 10, 64)
	log.Printf("Starting slangd %s built %s...\n", Version, time.Unix(buildTime, 0).Format(time.RFC3339))

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

func (e *EnvironPaths) loadDaemonServices(srv *daemon.Server) {
	srv.AddRedirect("/", "/app/")
	srv.AddAppServer("/app", http.Dir(e.SLANG_UI))
	srv.AddService("/operator", daemon.DefinitionService)
	srv.AddService("/run", daemon.RunnerService)
}

func (e *EnvironPaths) startDaemonServer(srv *daemon.Server) {
	url := fmt.Sprintf("http://%s:%d/", srv.Host, srv.Port)
	errors := make(chan error)
	go informUser(url, errors)
	errors <- srv.Run()
	select {} // prevent immidate exit when srv.Run failes --> informUser coroutine can handle error
}

func informUser(url string, errors chan error) {
	select {
	case err := <-errors:
		log.Fatal(fmt.Sprintf("\n\n\t%v\n\n", err))
	case <-time.After(500 * time.Millisecond):
		log.Printf("\n\n\tOpen following URL  %s  in your browser.\n\n", url)
		browser.OpenURL(url)
	}
	select {
	case err := <-errors:
		log.Fatal(fmt.Sprintf("\n\n\t%v\n\n", err))
	}
}
