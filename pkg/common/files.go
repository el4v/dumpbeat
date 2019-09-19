package common

import (
	"bufio"
	"dumpbeat/pkg/log"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
)

// MoveFile from source directory to destination directory
func MoveFile(src, dst string) error {
	err := copyFile(src, dst)
	if err != nil {
		return errors.Wrapf(err, "Error copy file from %s to %s", src, dst)
	}
	defer func() {
		err := os.Remove(src)
		if err != nil {
			log.Error(fmt.Sprintf("%s: Error file remove. %s", src, err.Error()))
		}
	}()
	return nil
}

func copyFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "Error stat file %s", src)
	}
	if !sfi.Mode().IsRegular() {
		return errors.Wrapf(err, "CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "Error stat destination file %s", dst)
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return errors.Wrapf(err, "CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return nil
		}
	}
	if err = os.Link(src, dst); err == nil {
		return nil
	}
	err = copyFileContents(src, dst)
	if err != nil {
		return errors.Wrapf(err, "Error copy file contents from %s to %s", src, dst)
	}
	return nil
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "Error open file %s", src)
	}
	defer func() {
		err := in.Close()
		if err != nil {
			log.Error(err.Error())
		}
	}()
	out, err := os.Create(dst)
	if err != nil {
		return errors.Wrapf(err, "Error create file %s", dst)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return errors.Wrapf(err, "Error copy from %s to %s", src, dst)
	}
	err = out.Sync()
	if err != nil {
		return errors.Wrapf(err, "Error sync %s", dst)
	}
	return nil
}

// ReadBigFile ...
func ReadBigFile(fileName string) ([]byte, error) {
	fi, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error open file %s", fileName)
	}
	defer func() {
		if err = fi.Close(); err != nil {
			log.Error(err.Error())
		}
	}()
	r := bufio.NewReader(fi)
	buf := make([]byte, 4096)
	_, err = r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "Error read first 4096 bytes from %s", fileName)
	}
	return buf, nil
}
