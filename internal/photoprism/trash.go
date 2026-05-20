package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TrashFolder is the directory name under originals where trashed files are stored.
// Names starting with "__" are ignored during indexing (see pkg/fs/name.go).
const TrashFolder = "__trash"

// TrashDir returns the absolute path to the trash folder in originals.
func TrashDir() string {
	return filepath.Join(Config().OriginalsPath(), TrashFolder)
}

// TrashDestination returns the absolute destination path for a file relative to originals.
func TrashDestination(relName string) string {
	return filepath.Join(TrashDir(), filepath.FromSlash(relName))
}

// MovePhotoToTrash moves all originals files of a photo to __trash and removes it from the index.
func MovePhotoToTrash(p *entity.Photo) (numMoved int, err error) {
	if p == nil {
		return 0, errors.New("photo is nil")
	}

	originalsPath := Config().OriginalsPath()

	if originalsPath == "" {
		return 0, errors.New("trash: originals path is empty")
	}

	trashDir := TrashDir()

	if err = fs.MkdirAll(trashDir); err != nil {
		return 0, fmt.Errorf("trash: %s", err)
	}

	p = p.PreloadFiles()
	files := p.AllFiles()
	moved := make(map[string]struct{})

	for _, file := range files {
		if file.FileRoot != entity.RootOriginals {
			continue
		}

		relName := filepath.ToSlash(file.FileName)

		if relName == "" || relName == TrashFolder || strings.HasPrefix(relName, TrashFolder+"/") {
			continue
		}

		srcPath := FileName(file.FileRoot, file.FileName)

		if _, ok := moved[srcPath]; ok {
			continue
		}

		n, moveErr := moveOriginalFileToTrash(srcPath, relName, moved)

		numMoved += n

		if moveErr != nil {
			return numMoved, moveErr
		}
	}

	yamlFileName, yamlRelName, yamlErr := p.YamlFileName(Config().OriginalsPath(), Config().SidecarPath())

	if yamlErr != nil {
		log.Warnf("photo: %s (trash %s)", yamlErr, clean.Log(yamlRelName))
	}

	indexFiles, delErr := p.DeletePermanently()

	if delErr != nil {
		return numMoved, delErr
	}

	// Remove YAML backup and ExifTool cache; do not delete moved originals.
	for _, file := range indexFiles {
		if exifJson, _ := ExifToolCacheName(file.FileHash); fs.FileExists(exifJson) {
			if rmErr := os.Remove(exifJson); rmErr != nil {
				log.Warnf("trash: failed to delete sidecar %s", clean.Log(filepath.Base(exifJson)))
			}
		}
	}

	if yamlFileName != "" && fs.FileExists(yamlFileName) {
		if rmErr := os.Remove(yamlFileName); rmErr != nil {
			log.Warnf("photo: failed to delete sidecar file %s", clean.Log(yamlRelName))
		}
	}

	return numMoved, nil
}

// moveOriginalFileToTrash moves a file from originals to __trash, preserving its relative path.
func moveOriginalFileToTrash(srcPath, relName string, moved map[string]struct{}) (numMoved int, err error) {
	if _, ok := moved[srcPath]; ok {
		return 0, nil
	}

	moved[srcPath] = struct{}{}

	if !fs.FileExists(srcPath) {
		log.Warnf("trash: skipped missing file %s", clean.Log(relName))
		return 0, nil
	}

	destPath := TrashDestination(relName)

	mediaFile, err := NewMediaFile(srcPath)

	if err != nil {
		return 0, err
	}

	if moveErr := mediaFile.Move(destPath, false); moveErr != nil {
		return 0, moveErr
	}

	numMoved++

	if n, sidecarErr := moveCoLocatedOriginalSidecars(srcPath, destPath, moved); sidecarErr != nil {
		return numMoved, sidecarErr
	} else {
		numMoved += n
	}

	log.Infof("trash: moved %s to %s", clean.Log(relName), clean.Log(filepath.Join(TrashFolder, relName)))

	return numMoved, nil
}

// moveCoLocatedOriginalSidecars moves unindexed sidecar files next to an original into trash.
func moveCoLocatedOriginalSidecars(srcPath, destPath string, moved map[string]struct{}) (numMoved int, err error) {
	srcDir := filepath.Dir(srcPath)
	destDir := filepath.Dir(destPath)
	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	originalsPath := Config().OriginalsPath()

	matches, globErr := filepath.Glob(filepath.Join(srcDir, base+".*"))

	if globErr != nil {
		return 0, globErr
	}

	for _, match := range matches {
		if match == srcPath {
			continue
		}

		if _, ok := moved[match]; ok {
			continue
		}

		sidecar, sidecarErr := NewMediaFile(match)

		if sidecarErr != nil {
			continue
		} else if !sidecar.IsSidecar() {
			continue
		}

		relName, relErr := filepath.Rel(originalsPath, match)

		if relErr != nil || relName == "" || strings.HasPrefix(relName, "..") {
			continue
		}

		relName = filepath.ToSlash(relName)
		destSidecar := filepath.Join(destDir, filepath.Base(match))

		if moveErr := sidecar.Move(destSidecar, false); moveErr != nil {
			return numMoved, moveErr
		}

		moved[match] = struct{}{}
		numMoved++

		log.Infof("trash: moved sidecar %s", clean.Log(relName))
	}

	return numMoved, nil
}
