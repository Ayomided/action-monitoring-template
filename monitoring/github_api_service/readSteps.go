// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// // Steps -> StepNumber(Order), JobID/StepID?, Name, Status(Completed || ), Conclusion(Success || Skipped || Failure)

// type StepOptions struct {
// 	CreatedAt   time.Time
// 	CompletedAt time.Time
// }

// type Step struct {
// 	Order    int
// 	StepName string
// 	Options  StepOptions
// }

// func CreateStep(order int, name string, options ...*StepOptions) *Step {
// 	var opts StepOptions
// 	if len(options) > 0 && options[0] != nil {
// 		// Use the provided StepOptions if it's not nil
// 		opts = *options[0]
// 	} else {
// 		// Use default StepOptions if none or nil is provided
// 		opts = StepOptions{
// 			CreatedAt:   time.Now(),
// 			CompletedAt: time.Time{},
// 		}
// 	}

// 	return &Step{
// 		Order:    order,
// 		StepName: name,
// 		Options:  opts,
// 	}
// }

// func parseSteps() {
// 	fd, err := os.ReadDir("runLogs/runner-2-jobRun-10431209862")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, f := range fd {
// 		if f.IsDir() {
// 			fdd, err := os.ReadDir("runLogs/runner-2-jobRun-10431209862/" + f.Name())
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			fmt.Println("THE FOLDER: ", f.Name())
// 			for _, fk := range fdd {
// 				fmt.Println(fk)
// 			}
// 		}
// 	}
// }

// func main() {
// 	// parseSteps()
// 	parseFileStartTime()
// 	// order, name := parseFileName()
// 	// fmt.Println(order, name)

// }

// func parseFileName() (int64, string) {
// 	name := "6_Install dependencies and run tests for Golang.txt"
// 	name, _ = strings.CutSuffix(name, ".txt")
// 	parsed := strings.SplitN(name, "_", 2)

// 	order, err := strconv.ParseInt(parsed[0], 10, 0)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	if len(parsed) == 2 {
// 		return order, parsed[1]
// 	}
// 	return 0, ""
// }

// func parseFileStartTime() {
// 	f, err := os.Open("runLogs/runner-2-jobRun-10431209862/build (golang)/1_Set up job.txt")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer f.Close()

// 	reader := bufio.NewReader(f)
// 	var firstLine, lastLine string
// 	for {
// 		line, err := reader.ReadString('\n')
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Fatalf("read file line error: %v", err)
// 		}
// 		if firstLine == "" {
// 			firstLine = line
// 		}
// 		lastLine = line
// 	}
// 	fmt.Printf("First line: %s\n", strings.Split(firstLine, " ")[0])
// 	fmt.Printf("Last line: %s", lastLine)
// }
