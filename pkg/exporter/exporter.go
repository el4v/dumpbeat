package exporter

import (
	"dumpbeat/pkg/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	CountUnprocessedFilesGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "dumpbeat",
			Name:      "count_files_in_dump_directory",
			Help:      "Count files in dump root folder",
		})
)

// StartExporter ...
func StartExporter(exporterPort int) error {
	prometheus.MustRegister(CountUnprocessedFilesGauge)
	if exporterPort < 1000 {
		log.Fatal(fmt.Sprintf("Expected port range 1000-65535. Given %d", exporterPort))
	}
	http.Handle("/metrics", promhttp.Handler())
	log.Info(fmt.Sprintf("Beginning to serve on port :%d. Path /metrics", exporterPort))
	err := http.ListenAndServe(fmt.Sprintf(":%d", exporterPort), nil)
	if err != nil {
		return err
	}
	return nil
}

func CountUnprocessedFilesGaugeHandler(rootDir string) {
	for {
		count := 0.0
		err := filepath.Walk(rootDir, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			count += 1
			return nil
		})
		if err != nil {
			CountUnprocessedFilesGauge.Set(-1)
		}
		CountUnprocessedFilesGauge.Set(count)
		<-time.After(60 * time.Second)
	}
}
