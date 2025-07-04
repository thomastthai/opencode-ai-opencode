package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipHidden(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{".git", true},
		{"/some/path/.git", true},
		{".idea", true},
		{"/some/path/.idea", true},
		{"node_modules", true},
		{"/some/path/node_modules", true},
		{"file.txt", false},
		{"/some/path/file.txt", false},
		{".hiddenfile", true},
		{"/some/path/.hiddenfile", true},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			actual := SkipHidden(tc.path)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
