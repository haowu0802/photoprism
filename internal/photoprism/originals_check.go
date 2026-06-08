package photoprism

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// ErrOriginalsUnavailable indicates originals storage cannot be accessed, for example
// when an external drive was disconnected. Index purge and cleanup must not run in
// this state to avoid clearing the database by mistake.
var ErrOriginalsUnavailable = errors.New("originals unavailable")

// VerifyOriginalsAccessible checks that originals storage is reachable before indexing,
// purge, or cleanup operations that could otherwise mark all files as missing.
func VerifyOriginalsAccessible(conf *config.Config) error {
	if conf == nil {
		return errors.New("config is nil")
	}

	path := conf.OriginalsPath()

	if path == "" {
		return fmt.Errorf("%w: path is empty", ErrOriginalsUnavailable)
	}

	if err := fs.DirAccessible(path); err != nil {
		return fmt.Errorf("%w: %s (%s)", ErrOriginalsUnavailable, path, err)
	}

	count, err := query.IndexedOriginalsFilesCount()

	if err != nil {
		return err
	}

	// Empty library without indexed files is fine.
	if count == 0 {
		return nil
	}

	// Indexed files exist but the directory is empty: disconnected mount or wrong path.
	if fs.DirIsEmpty(path) {
		return fmt.Errorf("%w: %s (%d indexed files, directory is empty)", ErrOriginalsUnavailable, path, count)
	}

	files, err := query.IndexedOriginalsFiles(5)

	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	accessible := 0

	for _, file := range files {
		if fs.FileExists(FileName(file.FileRoot, file.FileName)) {
			accessible++
		}
	}

	if accessible == 0 {
		return fmt.Errorf("%w: %s (none of %d sampled indexed files are accessible)", ErrOriginalsUnavailable, path, len(files))
	}

	return nil
}

// OriginalsUnavailableMessage returns a translated error message for clients and logs.
func OriginalsUnavailableMessage(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, ErrOriginalsUnavailable) {
		return i18n.Msg(i18n.ErrOriginalsUnavailable)
	}

	return err.Error()
}
