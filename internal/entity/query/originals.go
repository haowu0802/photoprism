package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// IndexedOriginalsFilesCount returns the number of non-missing originals files in the index.
func IndexedOriginalsFilesCount() (count int, err error) {
	err = entity.Db().
		Model(&entity.File{}).
		Where("file_root = ? AND file_missing = 0 AND deleted_at IS NULL", entity.RootOriginals).
		Count(&count).Error

	return count, err
}

// IndexedOriginalsFiles returns up to limit indexed originals files for storage checks.
func IndexedOriginalsFiles(limit int) (files entity.Files, err error) {
	if limit < 1 {
		limit = 1
	}

	err = entity.Db().
		Where("file_root = ? AND file_missing = 0 AND deleted_at IS NULL", entity.RootOriginals).
		Order("id ASC").
		Limit(limit).
		Find(&files).Error

	return files, err
}
