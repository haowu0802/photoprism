package photoprism

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestTrashPaths(t *testing.T) {
	c := config.TestConfig()
	c.Options().OriginalsPath = filepath.Join(c.CachePath(), "_tmp", "originals")

	dest := TrashDestination("2020/01/test.jpg")

	assert.Contains(t, dest, TrashFolder)
	assert.Contains(t, dest, "2020")
	assert.Equal(t, filepath.Join(c.OriginalsPath(), TrashFolder), TrashDir())
}
