package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/nxadm/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	duration     *prometheus.GaugeVec
	jobStartTime *prometheus.GaugeVec
	jobEndTime   *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		duration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "runner",
			Name:      "job_duration_in_seconds",
			Help:      "Duration of Github Action job",
		}, []string{"job_name"}),
		jobStartTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "runner",
			Name:      "job_start_time",
			Help:      "Start time of Github Action Job",
		}, []string{"job_name"}),
		jobEndTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "runner",
			Name:      "job_end_time",
			Help:      "End time of Github Action Job",
		}, []string{"job_name"}),
	}
	reg.MustRegister(m.duration, m.jobStartTime, m.jobEndTime)
	return m
}

var (
	startTime     float64
	endTime       float64
	jobStartTimes = make(map[string]float64)
	jobEndTimes   = make(map[string]float64)
)

func main() {
	var (
		promPort = flag.Int("prom.port", 9110, "port to expose prometheus metrics")
		logPath  = flag.String("runner.log", "/var/log/github-runner/runner.log", "path to runner log")
	)
	flag.Parse()

	registry := prometheus.NewRegistry()
	m := NewMetrics(registry)

	go tailRunnerLogs(m, *logPath)

	router := http.NewServeMux()
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	router.Handle("/metrics", promHandler)

	port := fmt.Sprintf(":%d", *promPort)
	log.Printf("Starting exporter on %q/metrics", port)
	err := http.ListenAndServe(port, router)
	if err != nil {
		log.Fatalf("Cannot start exporter server: %s", err)
	}
}

func tailRunnerLogs(m *metrics, logPath string) {
	t, err := tail.TailFile(logPath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Poll:      true})
	if err != nil {
		log.Printf("Error tailing file %s: %v", logPath, err)
		return
	}
	startPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}Z): Running job: (.+)`)
	endPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}Z): Job (.+) completed with result: (\S+)`)

	for line := range t.Lines {
		if line.Err != nil {
			log.Printf("Error reading line from %s: %v", logPath, line.Err)
			continue
		}

		if match := startPattern.FindStringSubmatch(line.Text); match != nil {
			timestamp := match[1]
			jobName := match[2]
			t, err := time.Parse("2006-01-02 15:04:05Z", timestamp)
			if err != nil {
				log.Printf("Error parsing time: %v", err)
				continue
			}
			startTime = float64(t.Unix())
			jobStartTimes[jobName] = startTime
			m.jobStartTime.With(prometheus.Labels{"job_name": jobName}).Set(startTime)
			log.Printf("Job started: %s at %s", jobName, timestamp)
		}

		if match := endPattern.FindStringSubmatch(line.Text); match != nil {
			timestamp := match[1]
			jobName := match[2]
			result := match[3]
			t, err := time.Parse("2006-01-02 15:04:05Z", timestamp)
			if err != nil {
				log.Printf("Error parsing time: %v", err)
				continue
			}
			endTime = float64(t.Unix())
			jobEndTimes[jobName] = endTime
			m.jobEndTime.With(prometheus.Labels{"job_name": jobName}).Set(endTime)

			if start, ok := jobStartTimes[jobName]; ok {
				duration := endTime - start
				m.duration.With(prometheus.Labels{"job_name": jobName}).Set(duration)
				log.Printf("Job completed: %s with result: %s at %s. Finished in: %e", jobName, result, timestamp, duration)
			} else {
				log.Printf("No start time found for job: %s", jobName)
			}
		}
	}
}
