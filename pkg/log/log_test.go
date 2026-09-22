package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestForOperatorKeepsConcurrentEventsSeparate(t *testing.T) {
	previous := logger
	defer func() { logger = previous }()
	var output bytes.Buffer
	base := logrus.New()
	base.SetOutput(&output)
	base.SetFormatter(&logrus.JSONFormatter{})
	logger = &stdLogger{logrus.NewEntry(base)}
	blueprintID := uuid.New()
	SetBlueprint(blueprintID, "blueprint")

	const count = 100
	identities := make(map[string]string, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		id, name := uuid.New(), fmt.Sprintf("operator-%d", i)
		identities[name] = id.String()
		wg.Add(1)
		go func() {
			defer wg.Done()
			ForOperator(id, name).Error(name)
		}()
	}
	wg.Wait()
	Error("runner event")
	decoder := json.NewDecoder(&output)
	for i := 0; i < count; i++ {
		var event map[string]interface{}
		require.NoError(t, decoder.Decode(&event))
		name := event["msg"].(string)
		require.Equal(t, name, event["operatorName"])
		require.Equal(t, identities[name], event["operatorId"])
		require.Equal(t, blueprintID.String(), event["blueprintId"])
		require.Equal(t, "blueprint", event["blueprintName"])
		delete(identities, name)
	}
	require.Empty(t, identities)
	var event map[string]interface{}
	require.NoError(t, decoder.Decode(&event))
	require.NotContains(t, event, "operatorId")
	require.NotContains(t, event, "operatorName")
}
