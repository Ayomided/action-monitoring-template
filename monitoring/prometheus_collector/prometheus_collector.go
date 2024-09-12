package prometheus_collector

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Metric struct {
	Name     string `json:"__name__"`
	Instance string `json:"instance"`
	Job      string `json:"job"`
	JobName  string `json:"job_name"`
}

type Result struct {
	Metric Metric          `json:"metric"`
	Value  [][]interface{} `json:"values"`
}

type Data struct {
	Result     []Result `json:"result"`
	ResultType string   `json:"resultType"`
}

type MetricData struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

var start time.Time
var end time.Time

func Collect(RUNNER_PROMETHEUS_URL string) ([]Result, error) {

	// Add Flag for start end and step
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, RUNNER_PROMETHEUS_URL+"/api/v1/query_range", nil)
	if err != nil {
		return nil, err
	}

	end = time.Now()
	start = time.Now().Add(-5 * time.Hour)
	q := request.URL.Query()
	q.Add("query", "runner_job_duration_in_seconds")

	q.Add("start", strconv.FormatInt(start.Unix(), 10))
	q.Add("end", strconv.FormatInt(end.Unix(), 10))
	q.Add("step", strconv.FormatInt(120, 10))

	request.URL.RawQuery = q.Encode()

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(responseBody))

	var runnerDuration MetricData
	err = json.Unmarshal(responseBody, &runnerDuration)
	if err != nil {
		return nil, err
	}

	// Extract the necessary fields
	var resultMetric []Result
	for _, result := range runnerDuration.Data.Result {
		resultMetric = append(resultMetric, Result{
			Metric{
				Name:     result.Metric.Name,
				Instance: result.Metric.Instance,
				Job:      result.Metric.Job,
				JobName:  result.Metric.JobName,
			}, result.Value,
		})
	}

	return resultMetric, nil
}

func Parse(metrics []Result) ([]string, []string, []string, []float64, []float64) {

	names := []string{}
	jobs := []string{}
	jobNames := []string{}
	timestamps := []float64{}
	values := []float64{}

	for _, metric := range metrics {
		for _, valuePair := range metric.Value {
			timestamp, ok1 := valuePair[0].(float64)
			valueStr, ok2 := valuePair[1].(string)
			if !ok1 || !ok2 {
				log.Fatalf("Unexpected value format: %v", metric.Value)
			}

			// Convert value to float64
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				log.Fatalf("Error converting value to float64: %v", err)
			}

			// Print with full precision
			names = append(names, metric.Metric.Name)
			jobs = append(jobs, metric.Metric.Job)
			jobNames = append(jobNames, metric.Metric.JobName)
			timestamps = append(timestamps, timestamp)
			values = append(values, value)
		}
	}
	return names, jobs, jobNames, timestamps, values
}
