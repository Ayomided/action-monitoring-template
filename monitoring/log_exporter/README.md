# Log Exporter

├── Dockerfile
├── go.mod
├── go.sum
├── log_exporter.go
├── README.md
├── runner.log
└── runner.sh

### The log exporter is written to work with Prometheus to expose the stdout logs from the GitHub Actions Runners

- *Dockerfile*
The Dockerfile builds the Log Exporter into an image for to be run when the GitHub Action Runner is active and running

- *log_exporter.go*
