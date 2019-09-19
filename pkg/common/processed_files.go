package common

import (
	"sync"
	"time"
)

// ProcessedFiles ...
type ProcessedFiles struct {
	Files map[string]time.Time
	mux   sync.Mutex
}

// Add file to queue
func (pf *ProcessedFiles) Add(fileName string, modifiedTime time.Time) {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	pf.Files[fileName] = modifiedTime
}

// Delete file from queue
func (pf *ProcessedFiles) Delete(filename string) {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	delete(pf.Files, filename)
}
