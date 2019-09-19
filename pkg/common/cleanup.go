package common

import (
	"dumpbeat/pkg/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CleanupEmptyFolders remove empty folders exclude root dir and project dir
func CleanupEmptyFolders(rootDir string) error {
	err := filepath.Walk(rootDir, func(fileName string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		if fileInfo.IsDir() {
			// Не удалять корневую директорию и директорию приложения
			if rootDir == fileName {
				return nil
			}
			tmp := strings.TrimPrefix(strings.Replace(fileName, rootDir, "", 1), string(os.PathSeparator))
			if len(strings.Split(tmp, string(os.PathSeparator))) == 1 {
				return nil
			}
			isEmpty, err := isEmptyDir(fileName)
			if err != nil {
				return errors.Wrapf(err, "Error check is empty dir for directory %s", fileName)
			}
			if isEmpty {
				err := os.Remove(fileName)
				if err != nil {
					return errors.Wrapf(err, "Error when remove not empty directory %s", fileName)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err.Error())
	}
	return nil
}

func isEmptyDir(name string) (bool, error) {
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return false, errors.Wrapf(err, "Error read directory %s", name)
	}
	return len(entries) == 0, nil
}
