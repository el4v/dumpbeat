package cmd

import (
	root "dumpbeat/pkg"
	"dumpbeat/pkg/common"
	"dumpbeat/pkg/consul"
	"dumpbeat/pkg/dump"
	"dumpbeat/pkg/exporter"
	"dumpbeat/pkg/log"
	"dumpbeat/pkg/version"
	"dumpbeat/pkg/watcher"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

var (
	config *root.Config
)

const (
	DumpDir             = "dump_dir"
	BackupDir           = "backup_dir"
	PatternFileFilter   = "pattern_file_filter"
	FileWaitTime        = "file_wait_time"
	APIUrl              = "api_url"
	APIToken            = "api_token"
	DaysToArchive       = "days_to_archive"
	NodeName            = "node_name"
	MaxFileSize         = "max_file_size"
	Aliases             = "aliases"
	LogLevel            = "log_level"
	ConsulHost          = "consul_host"
	ConsulServiceName   = "consul_service_name"
	ExporterBindAddress = "exporter_bind_address"
	ExporterBindPort    = "exporter_bind_port"
)

func init() {
	config = root.GetConfig()
	viper.SetEnvPrefix("DUMPBEAT")
	viper.AutomaticEnv()
	flags := rootCmd.Flags()
	flags.StringP(DumpDir, "", "/dumps", "Dumps directory")
	flags.StringP(BackupDir, "", "/backup-dumps", "Directory for backup dumps")
	flags.StringP(PatternFileFilter, "", "*.txt", "Pattern for dump search")
	flags.IntP(FileWaitTime, "", 900, "Time to wait file after create for send")
	flags.StringP(APIUrl, "", "", "Dump viewer API url")
	flags.StringP(APIToken, "", "", "Dump viewer API token")
	flags.IntP(DaysToArchive, "", 2, "Days to archive dumps")
	flags.StringP(NodeName, "", "", "Node name")
	flags.IntP(MaxFileSize, "", 15, "Max file size (Mb)")
	flags.StringP(Aliases, "", "", "Aliases for dumps app")
	flags.StringP(LogLevel, "", "info", "Log level (panic|fatal|error|warn|info|debug|trace)")
	flags.StringP(ConsulHost, "", "127.0.0.1:8500", "Consul host")
	flags.StringP(ConsulServiceName, "", "dumpbeat", "Consul service name")
	flags.StringP(ExporterBindAddress, "", config.ExporterBindAddress, "Exporter bind address")
	flags.IntP(ExporterBindPort, "", config.ExporterBindPort, "Exporter bind port")
	err := viper.BindPFlag(DumpDir, flags.Lookup(DumpDir))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(BackupDir, flags.Lookup(BackupDir))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(PatternFileFilter, flags.Lookup(PatternFileFilter))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(FileWaitTime, flags.Lookup(FileWaitTime))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(APIUrl, flags.Lookup(APIUrl))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(APIToken, flags.Lookup(APIToken))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(DaysToArchive, flags.Lookup(DaysToArchive))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(MaxFileSize, flags.Lookup(MaxFileSize))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(NodeName, flags.Lookup(NodeName))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(Aliases, flags.Lookup(Aliases))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(LogLevel, flags.Lookup(LogLevel))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(ConsulHost, flags.Lookup(ConsulHost))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(ConsulServiceName, flags.Lookup(ConsulServiceName))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(ExporterBindAddress, flags.Lookup(ExporterBindAddress))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = viper.BindPFlag(ExporterBindPort, flags.Lookup(ExporterBindPort))
	if err != nil {
		log.Fatal(err.Error())
	}
	rootCmd.Version = version.AsString()
}

var rootCmd = &cobra.Command{
	Use:   "dumpbeat",
	Short: "Dump worker",
	Long:  `Dump worker is utility for processing and send dumps from local machine to central dump server`,
	Run:   run,
	PreRun: func(cmd *cobra.Command, args []string) {
		config.DumpDir = viper.GetString(DumpDir)
		config.BackupDir = viper.GetString(BackupDir)
		config.PatternFileFilter = viper.GetString(PatternFileFilter)
		config.FileWaitTime = viper.GetInt(FileWaitTime)
		config.APIUrl = viper.GetString(APIUrl)
		config.APIToken = viper.GetString(APIToken)
		config.DaysToArchive = viper.GetInt(DaysToArchive)
		config.NodeName = viper.GetString(NodeName)
		config.MaxFileSize = viper.GetInt(MaxFileSize)
		config.Aliases = viper.GetString(Aliases)
		config.ConsulHost = viper.GetString(ConsulHost)
		config.ConsulServiceName = viper.GetString(ConsulServiceName)
		config.ExporterBindAddress = viper.GetString(ExporterBindAddress)
		config.ExporterBindPort = viper.GetInt(ExporterBindPort)
		config.LogLevel = viper.GetString(LogLevel)
		config.AliasesMap = make(map[string]string)
		if config.Aliases != "" {
			aliasesSlice := strings.Split(config.Aliases, ",")
			for _, alias := range aliasesSlice {
				tmpList := strings.Split(alias, ":")
				if len(tmpList) == 2 {
					config.AliasesMap[tmpList[0]] = tmpList[1]
				}
			}
		}
		if config.NodeName == "" {
			nodeName, err := os.Hostname()
			if err != nil {
				log.Error(fmt.Sprintf("%s. Error get hostname", err.Error()))
			}
			config.NodeName = nodeName
		}
		err := log.ConfigureLogging()
		if err != nil {
			log.Fatal(err.Error())
		}
	},
	Args:    nil,
	Version: "0.1.0",
}

func run(_ *cobra.Command, _ []string) {
	consulClient, err := consul.NewConsulClient(config.ConsulHost)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot register consul"))
	}
	go func() {
		err := exporter.StartExporter(config.ExporterBindPort)
		if err != nil {
			log.Fatal(err.Error())
		}
	}()
	go func() {
		exporter.CountUnprocessedFilesGaugeHandler(config.DumpDir)
	}()
	err = consulClient.Register(config.ConsulServiceName, config.ExporterBindAddress, config.ExporterBindPort)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Service %s registered in consul", config.ConsulServiceName))
	consulDeRegister := func() {
		err := consulClient.DeRegister(config.ConsulServiceName)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Info(fmt.Sprintf("Service %s de-registered in consul", config.ConsulServiceName))
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Info("Received an interrupt, stopping services...")
		consulDeRegister()
		os.Exit(0)
	}()
	defer func() {
		consulDeRegister()
	}()

	go watcher.FSWatch()
	for {
		err := filepath.Walk(config.DumpDir, dump.VisitFileWithWaitTime)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = common.CleanupEmptyFolders(config.DumpDir)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = common.ArchiveDumps(config.BackupDir, config.PatternFileFilter, config.DaysToArchive)
		if err != nil {
			log.Fatal(err.Error())
		}
		time.Sleep(time.Duration(120) * time.Second)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
