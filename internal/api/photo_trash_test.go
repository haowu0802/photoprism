package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrashPhoto(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		TrashPhoto(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/invalid/trash")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
