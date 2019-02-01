package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"time"

	"strconv"

	"os"

	"github.com/Bitspark/browser"
	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/utils"
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

var onlyDaemon bool
var skipChecks bool

func main() {
	flag.BoolVar(&onlyDaemon, "only-daemon", false, "Don't automatically open UI")
	flag.BoolVar(&skipChecks, "skip-checks", false, "Skip checking and updating UI and Lib")
	flag.Parse()

	buildTime, _ := strconv.ParseInt(BuildTime, 10, 64)
	if buildTime != 0 {
		log.Printf("Starting slangd %s built %s...\n", Version, time.Unix(buildTime, 0).Format(time.RFC3339))
		checkNewestVersion()
	} else {
		log.Println("Starting slangd (local build)...")
	}

	daemon.SlangVersion = Version

	envPaths := initEnvironPaths()

	srv := daemon.New(NewFileSystemStorage(), "localhost", PORT)

	if !skipChecks {
		envPaths.loadLocalComponents()
	}
	envPaths.loadDaemonServices(srv)
	envPaths.startDaemonServer(srv)
}

func initEnvironPaths() *EnvironPaths {
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	slangPath := filepath.Join(currUser.HomeDir, "slang")

	e := &EnvironPaths{
		slangPath,
		EnsureEnvironVar("SLANG_DIR", filepath.Join(slangPath, "projects")),
		EnsureEnvironVar("SLANG_LIB", filepath.Join(slangPath, "lib")),
		EnsureEnvironVar("SLANG_UI", filepath.Join(slangPath, "ui")),
	}
	if _, err = EnsureDirExists(e.SLANG_DIR); err != nil {
		log.Fatal(err)
	}
	if _, err = EnsureDirExists(e.SLANG_LIB); err != nil {
		log.Fatal(err)
	}
	if _, err = EnsureDirExists(e.SLANG_UI); err != nil {
		log.Fatal(err)
	}
	return e
}

func checkNewestVersion() {
	isNewest, newestVer, err := daemon.IsNewestSlangVersion(Version)
	if err != nil {
		log.Printf("Could not check newest slang version (%s)\n", err.Error())
		return
	}
	if isNewest {
		log.Printf("Your local slang is up-to-date (%s)\n", newestVer)
		return
	}
	log.Printf("Your local slang has version %v but latest is %v.", Version, newestVer)
	log.Printf("It is highly recommended to download the latest version.")
	log.Printf("Older versions might not be compatible with the newest slang-ui and slang-lib.")
	log.Printf("Do you want to download the newest slang version?")
	openBrowser := utils.AskForConfirmation("Invalid input")
	if openBrowser {
		browser.OpenURL("https://tryslang.com/#download")
		os.Exit(0)
	}
}

func (e *EnvironPaths) loadLocalComponents() {
	for repoName, dirPath := range map[string]string{"slang-lib": e.SLANG_LIB, "slang-ui": e.SLANG_UI} {
		dl := daemon.NewComponentLoaderLatestRelease(repoName, dirPath)
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

	// Load slang examples only when slang is started the first time
	if daemon.IsDirEmpty(e.SLANG_DIR) {
		dl := daemon.NewComponentLoaderLatestMaster("slang-examples", e.SLANG_DIR)
		log.Println("Downloading example operators.")
		if err := dl.Load(); err != nil {
			log.Fatal(err)
		}
		log.Printf("Done.")
	}

}

func (e *EnvironPaths) loadDaemonServices(srv *daemon.Server) {
	srv.AddRedirect("/", "/app/")
	srv.AddAppServer("/app", http.Dir(e.SLANG_UI))
	srv.AddAppServer("/studio", http.Dir(filepath.Join(filepath.Dir(e.SLANG_UI), "studio")))
	srv.AddService("/operator", daemon.DefinitionService)
	srv.AddService("/run", daemon.RunnerService)
	srv.AddService("/share", daemon.SharingService)
	srv.AddOperatorProxy("/instance")
}

func (e *EnvironPaths) startDaemonServer(srv *daemon.Server) {
	url := fmt.Sprintf("http://%s:%d/", srv.Host, srv.Port)
	errors := make(chan error)
	go informUser(url, errors)
	errors <- srv.Run()
	select {} // prevent immediate exit when srv.Run fails --> informUser goroutine can handle error
}

func informUser(url string, errors chan error) {
	select {
	case err := <-errors:
		log.Fatal(fmt.Sprintf("\n\n\t%v\n\n", err))
	case <-time.After(500 * time.Millisecond):
		log.Printf("\n\n\tOpen  %s  in your browser.\n\n", url)
		if !onlyDaemon {
			browser.OpenURL(url)
		}
	}
	select {
	case err := <-errors:
		log.Fatal(fmt.Sprintf("\n\n\t%v\n\n", err))
	}
}
