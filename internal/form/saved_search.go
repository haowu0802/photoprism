package form

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SavedSearch represents a saved search bookmark form.
type SavedSearch struct {
	Title    string `json:"Title"`
	Query    string `json:"Query"`
	Order    string `json:"Order"`
	Reverse  bool   `json:"Reverse"`
	Position int    `json:"Position"`
}

// Validate returns an error if any form values are invalid.
func (frm *SavedSearch) Validate() error {
	frm.Title = txt.Clip(strings.TrimSpace(frm.Title), txt.ClipLongName)

	if frm.Title == "" {
		return i18n.Error(i18n.ErrInvalidName)
	}

	frm.Query = strings.TrimSpace(frm.Query)

	if frm.Query == "" {
		return i18n.Error(i18n.ErrBadRequest)
	}

	search := &SearchPhotos{Query: frm.Query}

	if err := search.ParseQueryString(); err != nil {
		return err
	}

	frm.Order = strings.TrimSpace(frm.Order)

	switch frm.Order {
	case "", sortby.Random, sortby.Newest, sortby.Oldest:
	default:
		return i18n.Error(i18n.ErrBadRequest)
	}

	return nil
}
