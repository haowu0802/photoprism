package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// SearchSavedSearches returns saved searches for the authenticated user.
//
//	@Summary	returns saved searches for the authenticated user
//	@Id			SearchSavedSearches
//	@Tags		Saved Searches
//	@Produce	json
//	@Success	200				{array}		entity.SavedSearch
//	@Failure	401,403,429		{object}	i18n.Response
//	@Router		/api/v1/saved-searches [get]
func SearchSavedSearches(router *gin.RouterGroup) {
	router.GET("/saved-searches", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		result, err := query.SavedSearchesByUser(s.UserUID)

		if err != nil {
			log.Errorf("saved-search: %s", clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		AddCountHeader(c, len(result))
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}

// GetSavedSearch returns a saved search by UID.
//
//	@Summary	returns a saved search by UID
//	@Id			GetSavedSearch
//	@Tags		Saved Searches
//	@Produce	json
//	@Success	200					{object}	entity.SavedSearch
//	@Failure	401,403,404,429		{object}	i18n.Response
//	@Param		uid					path		string	true	"Saved Search UID"
//	@Router		/api/v1/saved-searches/{uid} [get]
func GetSavedSearch(router *gin.RouterGroup) {
	router.GET("/saved-searches/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		m, err := query.SavedSearchByUID(uid, s.UserUID)

		if err != nil {
			AbortNotFound(c)
			return
		}

		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, m)
	})
}

// CreateSavedSearch creates a new saved search for the authenticated user.
//
//	@Summary	creates a new saved search
//	@Id			CreateSavedSearch
//	@Tags		Saved Searches
//	@Accept		json
//	@Produce	json
//	@Success	201					{object}	entity.SavedSearch
//	@Failure	400,401,403,429		{object}	i18n.Response
//	@Param		savedSearch			body		form.SavedSearch	true	"saved search properties"
//	@Router		/api/v1/saved-searches [post]
func CreateSavedSearch(router *gin.RouterGroup) {
	router.POST("/saved-searches", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		var frm form.SavedSearch

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		} else if err = frm.Validate(); err != nil {
			AbortBadRequest(c, err)
			return
		}

		m := entity.NewSavedSearch(s.UserUID, frm.Title, frm.Query, frm.Order, frm.Reverse, frm.Position)

		if err := m.Create(); err != nil {
			log.Errorf("saved-search: %s (create)", clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		PublishSavedSearchEvent(StatusCreated, *m, c)

		header.SetLocation(c, c.FullPath(), m.SavedSearchUID)

		c.JSON(http.StatusCreated, m)
	})
}

// UpdateSavedSearch updates an existing saved search.
//
//	@Summary	updates a saved search
//	@Id			UpdateSavedSearch
//	@Tags		Saved Searches
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	entity.SavedSearch
//	@Failure	400,401,403,404,429	{object}	i18n.Response
//	@Param		uid					path		string				true	"Saved Search UID"
//	@Param		savedSearch			body		form.SavedSearch	true	"saved search properties"
//	@Router		/api/v1/saved-searches/{uid} [put]
func UpdateSavedSearch(router *gin.RouterGroup) {
	router.PUT("/saved-searches/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		m, err := query.SavedSearchByUID(uid, s.UserUID)

		if err != nil {
			AbortNotFound(c)
			return
		}

		var frm form.SavedSearch

		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err = c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		} else if err = frm.Validate(); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if err = m.SetForm(frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if err = m.Save(); err != nil {
			log.Errorf("saved-search: %s (update)", clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		PublishSavedSearchEvent(StatusUpdated, *m, c)

		c.JSON(http.StatusOK, m)
	})
}

// DeleteSavedSearch deletes a saved search.
//
//	@Summary	deletes a saved search
//	@Id			DeleteSavedSearch
//	@Tags		Saved Searches
//	@Produce	json
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"Saved Search UID"
//	@Router		/api/v1/saved-searches/{uid} [delete]
func DeleteSavedSearch(router *gin.RouterGroup) {
	router.DELETE("/saved-searches/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		m, err := query.SavedSearchByUID(uid, s.UserUID)

		if err != nil {
			AbortNotFound(c)
			return
		}

		if err = m.Delete(); err != nil && err != gorm.ErrRecordNotFound {
			log.Errorf("saved-search: %s (delete)", clean.Error(err))
			AbortDeleteFailed(c)
			return
		}

		PublishSavedSearchEvent(StatusDeleted, *m, c)

		c.JSON(http.StatusOK, gin.H{"status": StatusSuccess.String()})
	})
}

// PublishSavedSearchEvent publishes saved search changes to connected clients.
func PublishSavedSearchEvent(ev Event, m entity.SavedSearch, c *gin.Context) {
	event.PublishEntities("saved-searches", ev.String(), []entity.SavedSearch{m})
}
