package common

import (
	"archive/tar"
	"compress/gzip"
	"dumpbeat/pkg/log"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

func ArchiveDumps(rootDir, patternFileFilter string, daysToArchive int) error {
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	groupFiles, err := createFileListForArchive(rootDir, patternFileFilter, daysToArchive)
	if err != nil {
		return err
	}
	for day := range groupFiles {
		err := createTarball(path.Join(rootDir, fmt.Sprintf("%s.tar.gz", day)), groupFiles[day])
		if err != nil {
			log.Error(err.Error())
		}
	}
	return nil
}

func createFileListForArchive(rootDir string, patternFileFilter string, daysToArchive int) (map[string][]string, error) {
	groupFiles := make(map[string][]string)
	err := filepath.Walk(rootDir, func(fileName string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			log.Error(err.Error())
			return nil
		}
		if fileInfo.IsDir() {
			return nil
		}
		matched, err := filepath.Match(patternFileFilter, fileInfo.Name())
		if err != nil {
			log.Error(err.Error())
			return err
		}
		if matched {
			if int(time.Since(GetStartDay(fileInfo.ModTime())).Hours())/24 > daysToArchive {
				formattedDate := fileInfo.ModTime().Format("2006-01-02")
				groupFiles[formattedDate] = append(groupFiles[formattedDate], fileName)
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return groupFiles, nil
}

func createTarball(tarballFilePath string, filePaths []string) error {
	file, err := os.Create(tarballFilePath)
	if err != nil {
		return errors.Wrapf(err, "could not create tarball file %s", tarballFilePath)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error(fmt.Sprintf("%s: Error file close. %s", file.Name(), err.Error()))
		}
	}()

	gzipWriter := gzip.NewWriter(file)
	defer func() {
		err := gzipWriter.Close()
		if err != nil {
			log.Error(fmt.Sprintf("%s: Error gzip writer close. %s", file.Name(), err.Error()))
		}
	}()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		err := tarWriter.Close()
		if err != nil {
			log.Error(fmt.Sprintf("%s: Error tar writer close. %s", file.Name(), err.Error()))
		}
	}()

	for _, filePath := range filePaths {
		err := addFileToTarWriter(filePath, tarWriter)
		if err != nil {
			return errors.Wrapf(err, "could not add file %s, to tarball", filePath)
		}
		err = os.Remove(filePath)
		if err != nil {
			return errors.Wrapf(err, "Error remove %s after add file to tar")
		}
	}
	return nil
}

func addFileToTarWriter(filePath string, tarWriter *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.Wrapf(err, "could not open file %s", filePath)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error(fmt.Sprintf("%s: Error file close. %s", file.Name(), err.Error()))
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		return errors.Wrapf(err, "could not get stat for file %s", filePath)
	}

	header := &tar.Header{
		Name:    filePath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return errors.Wrapf(err, "could not write header for file %s", filePath)
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return errors.Wrapf(err, "could not copy the file %s data to the tarball", filePath)
	}

	return nil
}
