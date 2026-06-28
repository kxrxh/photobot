//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	testutil "csort.ru/coffeebot/internal/testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	if err := testutil.StartSharedInfra(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "start shared infra: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	testutil.StopSharedInfra()
	os.Exit(code)
}
