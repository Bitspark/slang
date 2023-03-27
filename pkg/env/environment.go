package env

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Bitspark/slang/pkg/utils"
)

// Holds configuration for the web parts
type httpCfg struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type Environment struct {
	SLANG_PATH          string
	SLANG_WORKSPACE     string
	SLANG_LIB_REPO_PATH string
	SLANG_LIB           string
	SLANG_UI            string

	HTTP httpCfg
}

func ensureEnvironVar(key string, dfltVal string) string {
	if val := os.Getenv(key); strings.Trim(val, " ") != "" {
		return val
	}
	os.Setenv(key, dfltVal)
	return dfltVal
}

func New(addr string, port int) *Environment {
	currUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	slangPath := filepath.Join(currUser.HomeDir, "slang")

	e := &Environment{
		slangPath,
		ensureEnvironVar("SLANG_DIR", filepath.Join(slangPath, "blueprints")),
		ensureEnvironVar("SLANG_LIB_REPO_PATH", filepath.Join(slangPath, "shared")),
		ensureEnvironVar("SLANG_LIB", filepath.Join(slangPath, "shared", "slang")),
		ensureEnvironVar("SLANG_UI", filepath.Join(slangPath, "ui")),
		httpCfg{Address: addr, Port: port},
	}

	if _, err = utils.EnsureDirExists(e.SLANG_WORKSPACE); err != nil {
		log.Fatal(err)
	}
	// we do not need to check the REPO as it will be present after
	// ensuring `SLANG_LIB` exists
	if _, err = utils.EnsureDirExists(e.SLANG_LIB); err != nil {
		log.Fatal(err)
	}
	if _, err = utils.EnsureDirExists(e.SLANG_UI); err != nil {
		log.Fatal(err)
	}

	return e
}
