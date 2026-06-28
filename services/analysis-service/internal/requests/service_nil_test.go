package requests

import (
	"context"
	"testing"

	"csort.ru/analysis-service/internal/transport/ws"
	"github.com/stretchr/testify/require"
)

func TestServiceBroadcastNilHubNoPanic(t *testing.T) {
	s := &Service{wsHub: nil}
	require.NotPanics(t, func() {
		s.broadcastToUser(context.Background(), "telegram:1", ws.Message{Type: "x"})
	})
}

func TestServiceDeleteTempNilClientNoPanic(t *testing.T) {
	s := &Service{tempStorage: nil}
	require.NotPanics(t, func() {
		s.deleteTempFiles(context.Background(), []string{"a", "b"})
	})
}

func TestServiceDeleteTempEmptyIDsNoPanic(t *testing.T) {
	s := &Service{tempStorage: nil}
	require.NotPanics(t, func() {
		s.deleteTempFiles(context.Background(), nil)
	})
}
