package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirAccessible(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		err := DirAccessible(filepath.Join(os.TempDir(), "does-not-exist-photoprism"))
		assert.Error(t, err)
	})
	t.Run("Success", func(t *testing.T) {
		err := DirAccessible("./testdata")
		assert.NoError(t, err)
	})
}
