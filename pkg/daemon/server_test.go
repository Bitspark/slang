package daemon

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHubBatchesMessages(t *testing.T) {
	hub := newHub()
	client := &ConnectedClient{userID: Root, send: make(chan []byte, 4)}
	go hub.run()
	hub.register <- client
	defer func() { hub.unregister <- client }()

	read := func(count int) []string {
		t.Helper()
		select {
		case data := <-client.send:
			var events []struct {
				Topic   string `json:"topic"`
				Payload string `json:"payload"`
			}
			require.NoError(t, json.Unmarshal(data, &events))
			require.Len(t, events, count)
			var payloads []string
			for _, event := range events {
				require.Equal(t, "Port", event.Topic)
				payloads = append(payloads, event.Payload)
			}
			return payloads
		case <-time.After(3 * time.Second):
			t.Fatal("hub did not deliver messages")
			return nil
		}
	}

	hub.broadCastTo(Root, Port, "first")
	hub.broadCastTo(Root, Port, "second")
	require.Equal(t, []string{"first", "second"}, read(2))
	// Once a batch has been delivered, later data starts a new batch.
	hub.broadCastTo(Root, Port, "third")
	require.Equal(t, []string{"third"}, read(1))
}
