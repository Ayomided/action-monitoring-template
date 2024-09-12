package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Ayomided/action-monitoring-template/monitoring/prometheus_collector"
	"github.com/subosito/gotenv"
)

func main() {
	if err := gotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	URL := os.Getenv("RUNNER_PROMETHEUS_URL")
	resp, err := http.Get(URL)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Prometheus Instance Failed to Respond\r\r\n Status %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stdout, "Prometheus Instance Failed to Respond\r\r\n Status %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	metric, err := prometheus_collector.Collect(URL)
	if err != nil {
		log.Fatalf("%v+: Error collecting metrics", err)
	}
	names, jobs, jobNames, timestamps, cpuValues := prometheus_collector.Parse(metric)

	csvFile, err := os.Create("metric.csv")
	if err != nil {
		log.Fatalf("Error creating file %s", err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	err = csvWriter.Write([]string{"Name", "Job", "JobName", "Timestamp", "Value"})
	if err == nil {
		log.Print("Successfully written to file")
	}
	defer csvWriter.Flush()

	for i := range jobNames {
		row := []string{names[i], jobs[i], jobNames[i], strconv.FormatFloat(timestamps[i], 'f', 0, 64), strconv.FormatFloat(cpuValues[i], 'f', 2, 64)}
		if err := csvWriter.Write(row); err != nil {
			log.Fatalf("Error writing metric row %s", err)
		}
	}

}
