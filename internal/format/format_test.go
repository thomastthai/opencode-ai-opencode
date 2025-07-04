package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatOutput(t *testing.T) {
	t.Run("text format", func(t *testing.T) {
		content := "hello world"
		formatted := FormatOutput(content, "text")
		assert.Equal(t, content, formatted)
	})

	t.Run("json format", func(t *testing.T) {
		content := "hello world"
		formatted := FormatOutput(content, "json")
		expected := `{
  "response": "hello world"
}`
		assert.JSONEq(t, expected, formatted)
	})
}
