package daemon

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func execEcho(typ core.TypeDef) core.Blueprint {
	return core.Blueprint{
		Id: uuid.New(), Meta: core.BlueprintMetaDef{Name: "Echo"},
		ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {In: typ, Out: typ}},
		Connections: map[string][]string{"(": {")"}},
	}
}

func newExecTestServer(t *testing.T, blueprints ...core.Blueprint) *httptest.Server {
	t.Helper()
	elem.Init()
	dir := t.TempDir()
	for _, key := range []string{"SLANG_DIR", "SLANG_LIB_REPO_PATH", "SLANG_LIB", "SLANG_UI"} {
		t.Setenv(key, filepath.Join(dir, key))
	}
	st := storage.NewStorage().AddBackend(storage.NewWritableFileSystem(dir))
	for _, bp := range blueprints {
		_, err := st.Save(bp)
		require.NoError(t, err)
	}
	previous := romanager
	romanager = &runningOperatorManager{
		ropByHandle: make(map[string]*runningOperator), handleByDefinition: make(map[string]string),
	}
	ctx := SetStorage(context.Background(), st)
	server := httptest.NewServer(NewServer(&ctx, env.New("localhost", 0), nil).Handler())
	server.Client().Timeout = 3 * time.Second
	t.Cleanup(func() {
		for _, rop := range romanager.List() {
			require.NoError(t, romanager.Halt(rop))
		}
		server.Close()
		romanager = previous
	})
	return server
}

func TestExecBlueprintPOST(t *testing.T) {
	number := execEcho(core.TypeDef{Type: "number"})
	text := execEcho(core.TypeDef{Type: "string"})
	generic := execEcho(core.TypeDef{Type: "generic", Generic: "value"})
	property := execEcho(core.TypeDef{Type: "number"})
	property.PropertyDefs = core.PropertyMap{"label": {Type: "string"}}
	server := newExecTestServer(t, number, text, generic, property)
	for _, tc := range []struct {
		name   string
		id     uuid.UUID
		body   string
		status int
		output string
	}{
		{"number input", number.Id, `{"input":21}`, http.StatusOK, `21`},
		{"reuse same definition", number.Id, `{"input":42}`, http.StatusOK, `42`},
		{"different blueprint with same properties", text.Id, `{"input":"hello"}`, http.StatusOK, `"hello"`},
		{"number generic", generic.Id, `{"generics":{"value":{"type":"number"}},"input":7}`, http.StatusOK, `7`},
		{"string generic with same properties", generic.Id, `{"generics":{"value":{"type":"string"}},"input":"world"}`, http.StatusOK, `"world"`},
		{"stream generic", generic.Id, `{"generics":{"value":{"type":"stream","stream":{"type":"number"}}},"input":[1,2]}`, http.StatusOK, `[1,2]`},
		{"required property supplied", property.Id, `{"properties":{"label":"example"},"input":12}`, http.StatusOK, `12`},
		{"required property missing", property.Id, `{"input":12}`, http.StatusBadRequest, ""},
		{"null is an output", number.Id, `{"input":null}`, http.StatusOK, `null`},
		{"wrong input type", number.Id, `{"input":"wrong"}`, http.StatusBadRequest, ""},
		{"malformed JSON", number.Id, `{`, http.StatusBadRequest, ""},
		{"trailing JSON", number.Id, `{"input":1} {"input":2}`, http.StatusBadRequest, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response, err := server.Client().Post(server.URL+"/run/"+tc.id.String()+"/", "application/json", strings.NewReader(tc.body))
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, tc.status, response.StatusCode, "%s", body)
			if tc.output != "" {
				require.JSONEq(t, tc.output, string(body))
			}
		})
	}
}

func TestRunningOperatorPullWaitsForDelayedAndNullOutputs(t *testing.T) {
	rop := &runningOperator{outgoing: make(chan interface{}), outStop: make(chan bool)}
	go func() {
		time.Sleep(600 * time.Millisecond)
		rop.outgoing <- 21
		rop.outgoing <- nil
		close(rop.outStop)
	}()
	value, ok := rop.Pull()
	require.True(t, ok, "a slow output must not be discarded after 500 ms")
	require.Equal(t, 21, value)
	value, ok = rop.Pull()
	require.True(t, ok, "null output must be distinguishable from a stopped operator")
	require.Nil(t, value)
	_, ok = rop.Pull()
	require.False(t, ok)
}

func TestDefinitionKeyIncludesPropertyNamesAndBoundaries(t *testing.T) {
	id := uuid.New()
	require.NotEqual(t,
		definitionKey(id, nil, core.Properties{"a": 1}),
		definitionKey(id, nil, core.Properties{"b": 1}))
	require.NotEqual(t,
		definitionKey(id, nil, core.Properties{"a": 1, "b": 23}),
		definitionKey(id, nil, core.Properties{"a": 12, "b": 3}))
	require.Equal(t,
		definitionKey(id, nil, core.Properties{"a": 1, "b": 2}),
		definitionKey(id, nil, core.Properties{"b": 2, "a": 1}))
}
