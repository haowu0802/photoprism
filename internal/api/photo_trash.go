package api

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// TrashPhoto moves a photo's originals to __trash and removes it from the index.
//
//	@Summary	moves a photo to trash and removes it from the index
//	@Id			TrashPhoto
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	i18n.Response
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/trash [post]
func TrashPhoto(router *gin.RouterGroup) {
	router.POST("/photos/:uid/trash", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Delete {
			AbortFeatureDisabled(c)
			return
		}

		uid := clean.UID(c.Param("uid"))
		p, err := query.PhotoPreloadByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		event.AuditWarn([]string{ClientIP(c), s.UserName, "trash", path.Join(p.PhotoPath, p.PhotoName+"*")})

		if _, trashErr := photoprism.MovePhotoToTrash(&p); trashErr != nil {
			log.Errorf("trash: %s", trashErr)
			AbortDeleteFailed(c)
			return
		}

		config.FlushUsageCache()
		entity.UpdateCountsAsync()
		UpdateClientConfig()

		event.EntitiesDeleted("photos", []string{p.PhotoUID})

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgMovedToTrash))
	})
}
