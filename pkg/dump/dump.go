package dump

import (
	"bytes"
	root "dumpbeat/pkg"
	"dumpbeat/pkg/common"
	"dumpbeat/pkg/log"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Dump struct
type Dump struct {
	Content         string       `json:"content" bson:"content"`
	Filename        string       `json:"filename" bson:"filename"`
	DateCreatedFile int32        `json:"date_created_file" bson:"date_created_file"`
	NodeName        string       `json:"node_name" bson:"node_name"`
	RootDir         string       `json:"root_dir" bson:"root_dir"`
	FileSize        int64        `json:"file_size" bson:"file_size"`
	BucketName      string       `json:"bucket_name" bson:"bucket_name"`
	Date            time.Time    `json:"date" bson:"date"`
	Config          *root.Config `json:"-"`
}

func (d Dump) apiUrl() string {
	config := root.GetConfig()
	appName := getAppName(d.Filename, d.RootDir)
	if alias, ok := config.AliasesMap[appName]; ok {
		appName = alias
	}
	return fmt.Sprintf("%s/%s/add", strings.TrimRight(d.Config.APIUrl, "/"), appName)
}

// SendDump to API
func (d Dump) SendDump() error {
	jsonValue, err := json.Marshal(d)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", d.apiUrl(), bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.Config.APIToken))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "Error send dump to API %s", d.Filename)
	}
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("dump not sended to %s %s", d.apiUrl(), response.Status)
	}
	return nil
}

// Move dump to backup directory
func (d Dump) Move(fileInfo os.FileInfo) (err error) {
	backupDir := d.Config.BackupDir + strings.Replace(filepath.Dir(d.Filename), d.Config.DumpDir, "", 1)
	err = os.MkdirAll(backupDir, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Error create dir %s", backupDir)
	}
	err = os.Rename(d.Filename, path.Join(backupDir, fileInfo.Name()))
	if err != nil {
		err = common.MoveFile(d.Filename, path.Join(backupDir, fileInfo.Name()))
		if err != nil {
			log.Error(err.Error())
			return errors.Wrapf(err, "Error move file %s to %s", d.Filename, path.Join(backupDir, fileInfo.Name()))
		}
	}
	return nil
}

func getAppName(filename, rootDir string) string {
	sep := string(os.PathSeparator)
	return strings.Trim(strings.Split(strings.TrimLeft(strings.Replace(filename, rootDir, "", 1), sep), sep)[0], sep)
}

func VisitFileWithoutWaitTime(fileName string, fileInfo os.FileInfo, err error) error {
	config := root.GetConfig()
	if err != nil {
		return nil
	}
	if fileInfo.IsDir() {
		return nil
	}
	matched, err := filepath.Match(config.PatternFileFilter, fileInfo.Name())
	if err != nil {
		return errors.Wrapf(err, "Error match pattern file filter in directory %s", fileInfo.Name())
	}
	if matched {
		log.Debug(fmt.Sprintf("Time to processing %s", fileName))
		err := processFile(fileName, fileInfo, false)
		if err != nil {
			log.Error(fmt.Sprintf("%s. Error process file %s", err.Error(), fileName))
			//! Посмотреть почему тут так сделано
			return nil
		}
	}
	return nil
}

func VisitFileWithWaitTime(fileName string, fileInfo os.FileInfo, err error) error {
	config := root.GetConfig()
	if err != nil {
		return nil
	}
	if fileInfo.IsDir() {
		return nil
	}
	matched, err := filepath.Match(config.PatternFileFilter, fileInfo.Name())
	if err != nil {
		return errors.Wrapf(err, "Error match pattern file filter in directory %s", fileInfo.Name())
	}
	if matched {
		if int(time.Since(fileInfo.ModTime()).Seconds()) < config.FileWaitTime {
			return nil
		}
		log.Debug(fmt.Sprintf("Time to processing %s", fileName))
		err := processFile(fileName, fileInfo, true)
		if err != nil {
			log.Error(fmt.Sprintf("%s. Error process file %s", err.Error(), fileName))
			//! Посмотреть почему тут так сделано
			return nil
		}
	}
	return nil
}

func processFile(fileName string, fileInfo os.FileInfo, backup bool) (err error) {
	config := root.GetConfig()

	var content []byte
	if fileInfo.Size() > int64(config.MaxFileSize*1048576) {
		log.Info(fmt.Sprintf("File %s exceeds maximum size %d.\n", fileName, config.MaxFileSize))
		content, err = common.ReadBigFile(fileName)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	} else {
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	year, month, _ := fileInfo.ModTime().Date()
	bucketName := fmt.Sprintf("bronevik-dumps-%d-%d", year, month)
	d := Dump{
		Content:         string(content),
		Filename:        fileName,
		DateCreatedFile: int32(fileInfo.ModTime().Unix()),
		NodeName:        config.NodeName,
		RootDir:         config.DumpDir,
		FileSize:        fileInfo.Size(),
		BucketName:      bucketName,
		Date:            time.Now(),
		Config:          config,
	}
	err = d.SendDump()
	if err != nil {
		log.Error(fmt.Sprintf("%s : Error send dump", err.Error()))
		return err
	}
	if backup {
		err = d.Move(fileInfo)
		if err != nil {
			log.Error(fmt.Sprintf("%s : Error move file: %s\n", err.Error(), fileName))
			return err
		}
	}
	return nil
}
