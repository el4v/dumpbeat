package root

// Config ...
type Config struct {
	DumpDir             string
	BackupDir           string
	PatternFileFilter   string
	FileWaitTime        int
	APIUrl              string
	APIToken            string
	DaysToArchive       int
	NodeName            string
	MaxFileSize         int
	SentryDSN           string
	Aliases             string
	ConsulHost          string
	ConsulServiceName   string
	ExporterBindAddress string
	ExporterBindPort    int
	LogLevel            string
	AliasesMap          map[string]string
}

var config Config

// GetConfig ...
func GetConfig() *Config {
	return &config
}
