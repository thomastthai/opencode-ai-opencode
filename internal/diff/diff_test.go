package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatDiff(t *testing.T) {
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
-hello
+world
 context
`
	formatted, err := FormatDiff(diff)
	assert.NoError(t, err)
	assert.NotEmpty(t, formatted)
}
