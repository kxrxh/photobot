//go:build integration

package integration

import (
	"context"
	"log"
	"os"
	"testing"

	"csort.ru/classification-service/integration/harness"
	"csort.ru/classification-service/internal/logger"
	"github.com/jarcoal/httpmock"
)

var sharedEnv *harness.TestEnv

func TestMain(m *testing.M) {
	logger.InitLogger()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	ctx := context.Background()
	env, cleanup, err := harness.SetupTestEnv(ctx)
	if err != nil {
		log.Fatalf("integration setup: %v", err)
	}
	sharedEnv = env
	defer cleanup()

	os.Exit(m.Run())
}

func testEnv(t *testing.T) *harness.TestEnv {
	t.Helper()
	if sharedEnv == nil {
		t.Fatal("shared integration env not initialized")
	}
	return sharedEnv
}
