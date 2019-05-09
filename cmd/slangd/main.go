package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"

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

var onlyDaemon bool
var skipChecks bool
var withoutUI bool

func main() {
	flag.BoolVar(&onlyDaemon, "only-daemon", false, "Don't automatically open UI")
	flag.BoolVar(&skipChecks, "skip-checks", false, "Skip checking and updating UI and Lib")
	flag.BoolVar(&withoutUI, "without-ui", false, "Do not serve the UI found in SLANG_UI")
	flag.Parse()

	buildTime, _ := strconv.ParseInt(BuildTime, 10, 64)
	if buildTime != 0 {
		log.Printf("Starting slangd %s built %s...\n", Version, time.Unix(buildTime, 0).Format(time.RFC3339))
		checkNewestVersion()
	} else {
		log.Println("Starting slangd (local build)...")
	}

	daemon.SlangVersion = Version

	env := env.New("localhost", PORT)

	if !skipChecks {
		loadLocalComponents(env)
	}

	st := storage.NewStorage().
		AddBackend(storage.NewWritableFileSystem(env.SLANG_DIR)).
		AddBackend(storage.NewReadOnlyFileSystem(env.SLANG_LIB))

	ctx := daemon.SetStorage(context.Background(), st)
	srv := daemon.NewServer(&ctx, env)

	if !withoutUI {
		srv.AddRedirect("/", "/app/")
		srv.AddStaticServer("/app", http.Dir(env.SLANG_UI))
	}

	startDaemonServer(srv)
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

func loadLocalComponents(e *env.Environment) {
	for repoName, dirPath := range map[string]string{"slang-lib": e.SLANG_LIB_REPO_PATH, "slang-ui": e.SLANG_UI} {
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
}

func startDaemonServer(srv *daemon.Server) {
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
		if !onlyDaemon && !withoutUI {
			log.Printf("\n\n\tOpen  %s  in your browser.\n\n", url)
			browser.OpenURL(url)
		}
	}
	select {
	case err := <-errors:
		log.Fatal(fmt.Sprintf("\n\n\t%v\n\n", err))
	}
}
