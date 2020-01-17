# Dumpbeat

## Description
Dump worker is utility for processing and send dumps from local machine to central dump server

## Usage
```
Usage:
  dumpbeat [flags]

Flags:
      --aliases string                 Aliases for dumps app
      --api_token string               Dump viewer API token
      --api_url string                 Dump viewer API url
      --backup_dir string              Directory for backup dumps (default "/backup-dumps")
      --consul_host string             Consul host (default "127.0.0.1:8500")
      --consul_service_name string     Consul service name (default "dumpbeat")
      --days_to_archive int            Days to archive dumps (default 2)
      --dump_dir string                Dumps directory (default "/dumps")
      --exporter_bind_address string   Exporter bind address
      --exporter_bind_port int         Exporter bind port
      --file_wait_time int             Time to wait file after create for send (default 900)
  -h, --help                           help for dumpbeat
      --log_level string               Log level (panic|fatal|error|warn|info|debug|trace) (default "info")
      --max_file_size int              Max file size (Mb) (default 15)
      --node_name string               Node name
      --pattern_file_filter string     Pattern for dump search (default "*.txt")
      --version                        version for dumpbeat
```
