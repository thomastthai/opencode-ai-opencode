package tools

import (
	"os"
	"testing"

	"github.com/opencode-ai/opencode/internal/config"
)

func TestMain(m *testing.M) {
	// Set up a dummy config for tests that need it.
	config.Init(config.Options{
		Version: "test",
	})
	os.Exit(m.Run())
}
