package watcher

import (
	root "dumpbeat/pkg"
	"dumpbeat/pkg/common"
	"dumpbeat/pkg/dump"
	"dumpbeat/pkg/log"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FSWatcher struct {
	*fsnotify.Watcher
}

func (fsWatcher FSWatcher) watch(pf *common.ProcessedFiles) {
	log.Info("Start filesystem watcher")
	for {
		select {
		case event, ok := <-fsWatcher.Events:
			if !ok {
				return
			}
			log.Debug("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					log.Error(err.Error())
					return
				}
				if !fileInfo.IsDir() {
					log.Debug("modified file:", event.Name)
					pf.Add(event.Name, time.Now())
				}
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				fileInfo, err := os.Stat(event.Name)
				if err != nil {
					log.Error(err.Error())
					return
				}
				if fileInfo.IsDir() {
					err := fsWatcher.Add(event.Name)
					if err != nil {
						log.Fatal(err.Error())
						return
					}
					log.Info(fmt.Sprintf("Added new directory %s for watch", event.Name))
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				err := fsWatcher.Remove(event.Name)
				if err != nil {
					log.Debug(err.Error())
					return
				}
				log.Info(fmt.Sprintf("Remove directory %s from watch", event.Name))
			}
		case err, ok := <-fsWatcher.Errors:
			if !ok {
				return
			}
			log.Error("error:", err)
		}
	}
}

func FSWatch() {
	config := root.GetConfig()
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	var fsWatcher FSWatcher
	fsWatcher.Watcher = w
	defer func() {
		err := fsWatcher.Close()
		if err != nil {
			log.Error(err.Error())
		}
	}()
	done := make(chan bool)
	pf := common.ProcessedFiles{}
	pf.Files = make(map[string]time.Time)
	go func() {
		for {
			if len(pf.Files) > 0 {
				for file, modifiedTime := range pf.Files {
					if time.Since(modifiedTime) > 5*time.Second {
						fileInfo, err := os.Stat(file)
						if err != nil {
							log.Error(fmt.Sprintf("%s: Error stat file %s.", err.Error(), file))
							return
						}
						err = dump.VisitFileWithoutWaitTime(file, fileInfo, nil)
						if err != nil {
							log.Error(fmt.Sprintf("%s. Error visit file %s", err.Error(), file))
							return
						}
						pf.Delete(file)
					}
				}
			}
			<-time.After(5 * time.Second)
		}
	}()
	go fsWatcher.watch(&pf)
	err = fsWatcher.Add(config.DumpDir)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Added %s to watch", config.DumpDir))
	objects, err := ioutil.ReadDir(config.DumpDir)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Add directories from root dir `%s` to watch", config.DumpDir))
	for _, object := range objects {
		if object.IsDir() {
			err := filepath.Walk(path.Join(config.DumpDir, object.Name()), func(fileName string, fileInfo os.FileInfo, err error) error {
				if err != nil {
					log.Error(err.Error())
					return nil
				}
				if fileInfo.IsDir() {
					err := fsWatcher.Add(fileName)
					if err != nil {
						log.Fatal(err.Error())
					}
					log.Info(fmt.Sprintf("Added %s to watch", fileName))
				}
				return nil
			})
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
	<-done
}
