package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Job struct {
	Name     string
	Steps    []Step
	Duration float64
}

type Step struct {
	Order     int64
	Name      string
	Duration  float64
	Timestamp time.Time
}

func main() {
	err := collectFiles()
	if err != nil {
		log.Fatalf("Error collecting files: %v", err)
	}

	jobs, err := processFiles()
	if err != nil {
		log.Fatalf("Error processing files: %v", err)
	}

	csvFile, err := os.Create("stepMetrics.csv")
	if err != nil {
		log.Fatalf("Error creating file %s", err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	err = csvWriter.Write([]string{"JobName", "StepName", "Order", "Timestamp", "Value"})
	if err == nil {
		log.Print("Successfully written to file")
	}
	defer csvWriter.Flush()

	for _, job := range jobs {
		for _, step := range job.Steps {
			row := []string{job.Name, step.Name, strconv.FormatInt(step.Order, 10), step.Timestamp.String(), strconv.FormatFloat(step.Duration, 'f', -1, 64)}
			if err := csvWriter.Write(row); err != nil {
				log.Fatalf("Error writing metric row %s", err)
			}
		}
	}
}

func processFiles() ([]Job, error) {
	var allJobs []Job
	runnerFolders, err := os.ReadDir("uncompressed")
	if err != nil {
		return nil, fmt.Errorf("no uncompressed folder: %w", err)
	}

	for _, runnerFolder := range runnerFolders {
		if !runnerFolder.IsDir() || !strings.HasPrefix(runnerFolder.Name(), "runner") {
			continue
		}

		runnerDir := filepath.Join("uncompressed", runnerFolder.Name())
		jobFolders, err := os.ReadDir(runnerDir)
		if err != nil {
			return nil, fmt.Errorf("error reading runner directory %s: %w", runnerDir, err)
		}

		for _, jobFolder := range jobFolders {
			if !jobFolder.IsDir() {
				continue
			}

			jobName := jobFolder.Name()
			jobDir := filepath.Join(runnerDir, jobName)
			stepFiles, err := os.ReadDir(jobDir)
			if err != nil {
				return nil, fmt.Errorf("error reading job directory %s: %w", jobDir, err)
			}

			var steps []Step
			var totalDuration float64

			for _, stepFile := range stepFiles {
				if stepFile.IsDir() {
					continue
				}
				filePath := filepath.Join(jobDir, stepFile.Name())
				step, err := processStep(filePath)
				if err != nil {
					return nil, fmt.Errorf("error processing step %s: %w", filePath, err)
				}

				steps = append(steps, step)
				totalDuration += step.Duration
			}

			outJob := Job{
				Name:     jobName,
				Steps:    steps,
				Duration: totalDuration,
			}
			allJobs = append(allJobs, outJob)
		}
	}

	return allJobs, nil
}

func processStep(filePath string) (Step, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Step{}, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		return Step{}, fmt.Errorf("file %s has insufficient lines", filePath)
	}

	firstLine := lines[0]
	lastLine := lines[len(lines)-1]
	if lastLine == "" && len(lines) > 1 {
		lastLine = lines[len(lines)-2]
	}

	startTime, err := time.Parse(time.RFC3339Nano, strings.Split(firstLine, " ")[0])
	if err != nil {
		return Step{}, fmt.Errorf("error parsing start time: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339Nano, strings.Split(lastLine, " ")[0])
	if err != nil {
		return Step{}, fmt.Errorf("error parsing end time: %w", err)
	}

	duration := endTime.Sub(startTime).Seconds()

	fileName := filepath.Base(filePath)
	stepName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	stepNameParts := strings.SplitN(stepName, "_", 2)

	stepOrder, err := strconv.ParseInt(stepNameParts[0], 10, 64)
	if err != nil {
		return Step{}, fmt.Errorf("error parsing step order: %w", err)
	}

	re := regexp.MustCompile(`^\d*_`)
	stepName = re.ReplaceAllString(stepName, "")

	return Step{
		Order:     stepOrder,
		Name:      stepName,
		Duration:  duration,
		Timestamp: endTime,
	}, nil
}

func collectFiles() error {
	zipFiles, err := filepath.Glob("dataset/runLogs/*.zip")
	if err != nil {
		return fmt.Errorf("error finding zip files: %w", err)
	}

	fmt.Println(len(zipFiles))

	var wg sync.WaitGroup
	errChan := make(chan error, len(zipFiles))

	for _, zipFile := range zipFiles {
		wg.Add(1)
		go func(zipFile string) {
			defer wg.Done()
			if err := performUnzip(zipFile); err != nil {
				errChan <- fmt.Errorf("error unzipping file %s: %w", zipFile, err)
			}
		}(zipFile)
	}
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func performUnzip(path string) error {
	// Derive destination directory name from zip file name
	zipFileName := filepath.Base(path)
	destDir := fmt.Sprintf("uncompressed/%s", strings.TrimSuffix(zipFileName, filepath.Ext(zipFileName)))

	// Unzip the file
	if err := unzip(path, destDir); err != nil {
		return fmt.Errorf("error unzipping file %s: %w", path, err)
	}

	fmt.Println("Successfully unzipped:", zipFileName)
	return nil
}

func unzip(zipFile, destDir string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	destDir = filepath.Clean(destDir)

	for _, file := range reader.File {
		if err := extractFile(file, destDir); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(file *zip.File, destDir string) error {
	destPath := filepath.Join(destDir, file.Name)
	destPath = filepath.Clean(destPath)

	// Check for file traversal attack
	if !strings.HasPrefix(destPath, destDir) {
		return fmt.Errorf("invalid file path: %s", file.Name)
	}

	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(destPath, file.Mode()); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return err
		}
	}

	return nil
}
