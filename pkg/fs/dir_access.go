package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// DirAccessible reports whether path exists, is a readable directory, and can be listed.
func DirAccessible(path string) error {
	if path == "" {
		return errors.New("directory path is empty")
	}

	info, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory not found")
		}

		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory")
	}

	f, err := os.Open(path) //nolint:gosec // path provided by caller

	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	if _, err = f.Readdirnames(1); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
