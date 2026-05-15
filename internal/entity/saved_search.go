package entity

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	SavedSearchUID = byte('b')
)

var savedSearchMutex = sync.Mutex{}

// SavedSearches groups saved search bookmarks.
type SavedSearches []SavedSearch

// SavedSearch represents a user-defined photo search bookmark.
type SavedSearch struct {
	SavedSearchUID string     `gorm:"type:VARBINARY(42);primary_key;" json:"UID" yaml:"UID,omitempty"`
	UserUID        string     `gorm:"type:VARBINARY(42);index:idx_saved_searches_user;" json:"-" yaml:"-"`
	SearchTitle    string     `gorm:"type:VARCHAR(200);" json:"Title" yaml:"Title,omitempty"`
	SearchQuery    string     `gorm:"type:VARBINARY(2048);" json:"Query" yaml:"Query,omitempty"`
	SearchOrder    string     `gorm:"type:VARBINARY(32);" json:"Order" yaml:"Order,omitempty"`
	SearchReverse  bool       `json:"Reverse" yaml:"Reverse,omitempty"`
	SearchPosition int        `json:"Position" yaml:"Position,omitempty"`
	CreatedAt      time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt      time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt      *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity table name.
func (SavedSearch) TableName() string {
	return "saved_searches"
}

// BeforeCreate assigns a random UID before inserting a new row.
func (m *SavedSearch) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUnique(m.SavedSearchUID, SavedSearchUID) {
		return nil
	}

	return scope.SetColumn("SavedSearchUID", rnd.GenerateUID(SavedSearchUID))
}

// NewSavedSearch returns a new saved search entity for the specified user.
func NewSavedSearch(userUID, title, query, order string, reverse bool, position int) *SavedSearch {
	now := Now()

	return &SavedSearch{
		SavedSearchUID: rnd.GenerateUID(SavedSearchUID),
		UserUID:        userUID,
		SearchTitle:    txt.Clip(title, txt.ClipLongName),
		SearchQuery:    txt.Clip(query, 2048),
		SearchOrder:    txt.Clip(order, 32),
		SearchReverse:  reverse,
		SearchPosition: position,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Create inserts the entity into the database.
func (m *SavedSearch) Create() error {
	savedSearchMutex.Lock()
	defer savedSearchMutex.Unlock()

	return Db().Create(m).Error
}

// Save persists the entity.
func (m *SavedSearch) Save() error {
	savedSearchMutex.Lock()
	defer savedSearchMutex.Unlock()

	m.UpdatedAt = Now()

	return Db().Save(m).Error
}

// Delete soft-deletes the entity.
func (m *SavedSearch) Delete() error {
	savedSearchMutex.Lock()
	defer savedSearchMutex.Unlock()

	return Db().Delete(m).Error
}

// SetForm copies validated form values into the entity.
func (m *SavedSearch) SetForm(frm form.SavedSearch) error {
	m.SearchTitle = txt.Clip(frm.Title, txt.ClipLongName)
	m.SearchQuery = txt.Clip(frm.Query, 2048)
	m.SearchOrder = txt.Clip(frm.Order, 32)
	m.SearchReverse = frm.Reverse
	m.SearchPosition = frm.Position

	return nil
}
