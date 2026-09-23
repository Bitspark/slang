package elem

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Condition slang.elem.host-access-through-capabilities: no operator reaches the
// host except through the Capabilities its host provides.

// hostAccess maps an import path to the identifiers through which it reaches the
// host, or to "*" when any use of the package does. It names known entry points;
// it is a guard against regressions, not a proof that no other route exists.
var hostAccess = map[string][]string{
	"os": {"Open", "OpenFile", "Create", "ReadFile", "WriteFile", "Remove", "RemoveAll",
		"Mkdir", "MkdirAll", "Rename", "Stat", "Lstat", "ReadDir", "Truncate", "Chmod",
		"Chown", "Symlink", "Link", "CreateTemp", "MkdirTemp", "TempDir", "DirFS",
		"Getenv", "LookupEnv", "Environ", "Hostname", "Getwd", "Chdir", "StartProcess"},
	"io/ioutil":     {"ReadFile", "WriteFile", "ReadDir", "TempFile", "TempDir"},
	"path/filepath": {"Walk", "WalkDir", "Glob", "Abs", "EvalSymlinks"},
	"archive/zip":   {"OpenReader"},
	"os/exec":       {"*"},
	"net": {"Dial", "DialTimeout", "DialTCP", "DialUDP", "DialIP", "DialUnix", "Dialer",
		"Listen", "ListenTCP", "ListenUDP", "ListenIP", "ListenUnix", "ListenPacket",
		"ListenConfig", "LookupHost", "LookupIP", "LookupAddr", "LookupCNAME", "LookupMX"},
	"net/http": {"DefaultClient", "DefaultTransport", "Get", "Post", "PostForm", "Head",
		"ListenAndServe", "ListenAndServeTLS", "Serve", "ServeTLS", "Server", "Transport"},
	"net/smtp":                            {"*"},
	"database/sql":                        {"*"},
	"github.com/go-redis/redis":           {"*"},
	"github.com/Shopify/sarama":           {"*"},
	"github.com/eclipse/paho.mqtt.golang": {"*"},
	"time": {"Now", "Since", "Until", "Sleep", "After", "AfterFunc", "Tick",
		"NewTimer", "NewTicker"},
}

// packageNames gives the declared name of watched packages whose name differs from
// the last element of their import path.
var packageNames = map[string]string{"github.com/eclipse/paho.mqtt.golang": "mqtt"}

// capabilityAdapters implement the capabilities, so they may reach the host.
var capabilityAdapters = map[string]bool{"capabilities.go": true}

// reachesHostDirectly lists the operators not yet moved behind a capability, by
// the import paths they reach the host through. It may only shrink: no file may
// start reaching the host, and a file that stops must be removed here.
var reachesHostDirectly = map[string][]string{
	"data_variable_get.go":        {"time"},
	"database_execute.go":         {"database/sql"},
	"database_kafka_subscribe.go": {"github.com/Shopify/sarama"},
	"database_query.go":           {"database/sql"},
	"database_redis_get.go":       {"github.com/go-redis/redis"},
	"database_redis_hget.go":      {"github.com/go-redis/redis"},
	"database_redis_hincrby.go":   {"github.com/go-redis/redis"},
	"database_redis_hset.go":      {"github.com/go-redis/redis"},
	"database_redis_lpush.go":     {"github.com/go-redis/redis"},
	"database_redis_set.go":       {"github.com/go-redis/redis"},
	"database_redis_subscribe.go": {"github.com/go-redis/redis"},
	"files_append.go":             {"os"},
	"files_read.go":               {"io/ioutil"},
	"files_read_lines.go":         {"os"},
	"files_write.go":              {"io/ioutil"},
	"net_http_server.go":          {"net/http"},
	"net_mqtt_publish.go":         {"github.com/eclipse/paho.mqtt.golang"},
	"net_mqtt_subscribe.go":       {"github.com/eclipse/paho.mqtt.golang"},
	"net_send_email.go":           {"net/smtp"},
	"rand_range.go":               {"time"},
	"shell_execute.go":            {"os/exec"},
	"time_date_now.go":            {"time"},
	"time_delay.go":               {"time"},
	"time_unix.go":                {"time"},
}

// hostAccessIn returns the watched import paths src reaches the host through, each
// with the first identifier found.
func hostAccessIn(filename string, src interface{}) (map[string]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filename, src, 0)
	if err != nil {
		return nil, err
	}
	local := map[string]string{}
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if _, watched := hostAccess[path]; !watched {
			continue
		}
		name := path[strings.LastIndex(path, "/")+1:]
		if declared, ok := packageNames[path]; ok {
			name = declared
		}
		if imp.Name != nil {
			name = imp.Name.Name
		}
		local[name] = path
	}
	found := map[string]string{}
	ast.Inspect(file, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Obj != nil { // Obj is set when a local name shadows the package
			return true
		}
		path, ok := local[pkg.Name]
		if !ok {
			return true
		}
		for _, symbol := range hostAccess[path] {
			if symbol == "*" || symbol == sel.Sel.Name {
				if _, seen := found[path]; !seen {
					found[path] = pkg.Name + "." + sel.Sel.Name
				}
			}
		}
		return true
	})
	return found, nil
}

func TestOperatorsReachHostOnlyThroughCapabilities(t *testing.T) {
	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]map[string]string{}
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") || capabilityAdapters[name] {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		access, err := hostAccessIn(name, src)
		if err != nil {
			t.Fatal(err)
		}
		if len(access) > 0 {
			found[name] = access
		}
	}

	var problems []string
	for file, access := range found {
		allowed := map[string]bool{}
		for _, path := range reachesHostDirectly[file] {
			allowed[path] = true
		}
		for path, symbol := range access {
			if !allowed[path] {
				problems = append(problems, file+" reaches the host through "+symbol+
					" ("+path+"); take it from Capabilities instead")
			}
		}
	}
	for file, paths := range reachesHostDirectly {
		for _, path := range paths {
			if _, still := found[file][path]; !still {
				problems = append(problems, file+" no longer reaches the host through "+path+
					"; remove it from reachesHostDirectly")
			}
		}
	}
	sort.Strings(problems)
	for _, problem := range problems {
		t.Error(problem)
	}
}

// Control for the guard above: it must see host access, including through an
// aliased import, and must ignore a local name that shadows a package.
func TestHostAccessGuardDetectsDirectAccess(t *testing.T) {
	src := `package elem
import (
	disk "os"
	"net/http"
	"time"
)
func f(time struct{ Now func() }) {
	disk.ReadFile("x")
	http.Get("http://example.com")
	time.Now()
}`
	found, err := hostAccessIn("sample.go", src)
	if err != nil {
		t.Fatal(err)
	}
	if found["os"] != "disk.ReadFile" || found["net/http"] != "http.Get" {
		t.Errorf("guard missed direct access: %v", found)
	}
	if _, shadowed := found["time"]; shadowed {
		t.Errorf("guard reported a local variable named like a package: %v", found)
	}
}
