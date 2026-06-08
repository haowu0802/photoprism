package photoprism

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestVerifyOriginalsAccessible_EmptyPath(t *testing.T) {
	c := config.TestConfig()
	c.Options().OriginalsPath = ""

	err := VerifyOriginalsAccessible(c)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrOriginalsUnavailable))
}

func TestVerifyOriginalsAccessible_MissingDirectoryWithIndexedFiles(t *testing.T) {
	c := config.TestConfig()
	tmp := filepath.Join(c.CachePath(), "_tmp", "TestVerifyOriginalsAccessible")

	if err := fs.MkdirAll(tmp); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	missing := filepath.Join(tmp, "missing-originals")
	c.Options().OriginalsPath = missing

	// Use non-empty path that does not exist; indexed file count is zero in test DB.
	err := VerifyOriginalsAccessible(c)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrOriginalsUnavailable))
}
