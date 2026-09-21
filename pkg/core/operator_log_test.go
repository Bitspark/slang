package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type eventWriter chan []byte

func (w eventWriter) Write(p []byte) (int, error) {
	w <- append([]byte(nil), p...)
	return len(p), nil
}

func TestOperatorPanicIncludesStructuredIdentity(t *testing.T) {
	previous := logrus.StandardLogger().Out
	defer logrus.SetOutput(previous)
	output := make(eventWriter, 1)
	logrus.SetOutput(output)
	id := uuid.New()
	o, err := NewOperator("failing-operator", func(*Operator) {
		panic("expected map")
	}, nil, nil, nil, Blueprint{Id: id})
	require.NoError(t, err)
	o.Start()
	select {
	case data := <-output:
		var event map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &event))
		require.Equal(t, "failing-operator", event["operatorName"])
		require.Equal(t, id.String(), event["operatorId"])
		require.Equal(t, "operator panicked: expected map", event["msg"])
	case <-time.After(time.Second):
		t.Fatal("missing panic log event")
	}
	o.WaitForStop()
}
