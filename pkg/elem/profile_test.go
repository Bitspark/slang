package elem

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Conditions slang.builtins.requirements-agree and slang-infra.runner.safe-mode.

func TestResolveFollowsThePrecedenceOfStates(t *testing.T) {
	nothing := Profile{Name: "nothing", Version: 1}
	forbidding := Profile{Name: "forbidding", Version: 1, Implemented: Operations}
	httpClient := netHTTPClientCfg.blueprint.Id

	require.Equal(t, Available, LocalProfile.Resolve(httpClient).State)
	require.Equal(t, Available, nothing.Resolve(dataValueCfg.blueprint.Id).State, "a pure built-in needs nothing")

	unavailable := nothing.Resolve(httpClient)
	require.Equal(t, Unavailable, unavailable.State)
	require.EqualError(t, unavailable.Err(), `built-in "HTTP client" is unavailable under profile nothing/1: http.request is not implemented`)

	denied := forbidding.Resolve(httpClient)
	require.Equal(t, Denied, denied.State)
	require.EqualError(t, denied.Err(), `built-in "HTTP client" is denied under profile forbidding/1: http.request is not permitted`)

	removed := LocalProfile.Resolve(uuid.MustParse("362482c1-2021-4e5c-9463-b580a6c1967e"))
	require.Equal(t, Removed, removed.State)
	require.EqualError(t, removed.Err(), `built-in "Redis Get" was removed after v0.1.27`)

	unknown := LocalProfile.Resolve(uuid.New())
	require.Equal(t, Unknown, unknown.State)
	require.Error(t, unknown.Err())
}

func TestResolveReportsEveryCause(t *testing.T) {
	cfg := &builtinConfig{blueprint: dataValueCfg.blueprint.Copy(true),
		requires: []Requirement{{Operation: FilesRead}, {Operation: HTTPRequest}}}
	cfg.blueprint.Id, cfg.blueprint.Meta.Name = uuid.New(), "two effects"
	cfgs[cfg.blueprint.Id] = cfg
	t.Cleanup(func() { delete(cfgs, cfg.blueprint.Id) })

	mixed := Profile{Name: "mixed", Version: 1, Implemented: []Operation{HTTPRequest}}
	r := mixed.Resolve(cfg.blueprint.Id)
	require.Equal(t, Unavailable, r.State, "unavailable takes precedence over denied")
	require.Equal(t, []Cause{{FilesRead, Unavailable}, {HTTPRequest, Denied}}, r.Causes)
}

// The hosted runtime has no file namespace yet, so a hosted program can neither
// write nor read files; it keeps computation and HTTP.
func TestHostedProfileLeavesFileEffectsUnavailable(t *testing.T) {
	for _, cfg := range []*builtinConfig{filesWriteCfg, filesAppendCfg, filesReadCfg, filesReadLinesCfg} {
		require.Equal(t, Unavailable, HostedProfile.Resolve(cfg.blueprint.Id).State, cfg.blueprint.Meta.Name)
	}
	for _, cfg := range []*builtinConfig{netHTTPClientCfg, dataEvaluateCfg, filesZIPPackCfg, timeDateNowCfg} {
		require.Equal(t, Available, HostedProfile.Resolve(cfg.blueprint.Id).State, cfg.blueprint.Meta.Name)
	}
}

func TestPublicProfileDeniesEveryEffectButClockRandomnessAndState(t *testing.T) {
	for _, entry := range PublicProfile.Catalog() {
		if entry.State == Removed {
			continue
		}
		want := Available
		for _, req := range entry.Requirements {
			if !contains(PublicProfile.Permitted, req.Operation) {
				want = Denied
			}
		}
		require.Equal(t, want, entry.State, entry.Name)
	}
	require.Equal(t, Denied, PublicProfile.Resolve(netSendEmailCfg.blueprint.Id).State)
}

func TestCatalogListsEveryBuiltinOnce(t *testing.T) {
	catalog := LocalProfile.Catalog()
	require.Len(t, catalog, len(cfgs)+len(removedBuiltins))
	seen := map[uuid.UUID]bool{}
	for _, entry := range catalog {
		require.False(t, seen[entry.ID], entry.Name)
		seen[entry.ID] = true
		require.NotEqual(t, Unknown, entry.State, entry.Name)
	}
}

// A built-in that reaches the host must declare what it does there: the guard's
// list of files that still reach it directly, and the HTTP client, which reaches
// it through its capability, each need a requirement.
func TestEveryBuiltinThatReachesTheHostDeclaresItsRequirement(t *testing.T) {
	ids := builtinIdsByFile(t)
	reaching := map[string]bool{"net_http_client.go": true}
	for file := range reachesHostDirectly {
		reaching[file] = true
	}
	for file := range reaching {
		id, ok := ids[file]
		require.True(t, ok, "%s defines no built-in", file)
		require.NotEmpty(t, cfgs[id].requires, "%s reaches the host but declares no requirement", file)
	}
	require.Equal(t, []Requirement{{Operation: HTTPRequest, Dynamic: true}}, netHTTPClientCfg.requires)
}

// builtinIdsByFile maps each source file that defines a built-in to its ID.
func builtinIdsByFile(t *testing.T) map[string]uuid.UUID {
	t.Helper()
	names, err := filepath.Glob("*.go")
	require.NoError(t, err)
	ids := map[string]uuid.UUID{}
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		require.NoError(t, err)
		file, err := parser.ParseFile(token.NewFileSet(), name, src, 0)
		require.NoError(t, err)
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			if ident, ok := lit.Type.(*ast.Ident); !ok || ident.Name != "builtinConfig" {
				return true
			}
			for _, cfg := range cfgs {
				if strings.Contains(string(src), `"`+cfg.blueprint.Id.String()+`"`) {
					ids[name] = cfg.blueprint.Id
				}
			}
			return false
		})
	}
	return ids
}
