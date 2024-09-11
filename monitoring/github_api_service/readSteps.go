package main

import (
	"fmt"
	"log"
	"os"
)

func parseSteps() {
	fd, err := os.ReadDir("runLogs/runner-2-jobRun-10431209862")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fd {
		fmt.Println(f)
	}
}
