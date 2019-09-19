package main

import (
	"dumpbeat/cmd"
	"dumpbeat/pkg/log"
)

func main() {
	log.AddFields(map[string]interface{}{"App": "dumpbeat"}).Info("Start application")
	cmd.Execute()
}
