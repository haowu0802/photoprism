package query

import (
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
)

// SavedSearchesByUser returns all saved searches owned by a user.
func SavedSearchesByUser(userUID string) (entity.SavedSearches, error) {
	result := make(entity.SavedSearches, 0)

	if userUID == "" {
		return result, nil
	}

	err := Db().
		Where("user_uid = ?", userUID).
		Order("search_position ASC, search_title ASC").
		Find(&result).Error

	return result, err
}

// SavedSearchByUID returns a saved search if it exists and belongs to the user.
func SavedSearchByUID(uid, userUID string) (*entity.SavedSearch, error) {
	result := &entity.SavedSearch{}

	if uid == "" || userUID == "" {
		return result, gorm.ErrRecordNotFound
	}

	err := Db().
		Where("saved_search_uid = ? AND user_uid = ?", uid, userUID).
		First(result).Error

	return result, err
}
