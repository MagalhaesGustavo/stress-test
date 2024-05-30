package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	url         string
	requests    int
	concurrency int
)

var rootCmd = &cobra.Command{
	Use:   "stress-test",
	Short: "stress-test is a simple CLI tool to test the performance of a web server",

	PreRun: func(cmd *cobra.Command, args []string) {
		if url == "" {
			fmt.Println("URL is required")
			cmd.Help()
			os.Exit(1)
		}
		if requests < concurrency {
			fmt.Println("Number of requests should be greater than concurrency")
			cmd.Help()
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		processStressTest(url, requests, concurrency)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "URL to test")
	rootCmd.PersistentFlags().IntVarP(&requests, "requests", "r", 10, "Number of requests to perform")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of multiple requests to make at a time")

	rootCmd.MarkFlagsRequiredTogether("requests", "url")
}

type Report struct {
	totalRequests int
	status200     int
	statusOther   map[int]int
	totalTime     time.Duration
}

func processStressTest(url string, requests int, concurrency int) {
	fmt.Printf("Starting stress test on %s with %d requests and %d concurrency\n", url, requests, concurrency)

	report := &Report{
		statusOther: make(map[int]int),
	}

	startTime := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()
			makeRequest(url, report)
			<-semaphore
		}()
	}

	wg.Wait()
	report.totalTime = time.Since(startTime)

	printReport(report)
}

func makeRequest(url string, report *Report) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error making the request: %v", err)
		return
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading the response body: %v", err)
		return
	}

	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	report.totalRequests++
	if resp.StatusCode == http.StatusOK {
		report.status200++
	} else {
		report.statusOther[resp.StatusCode]++
	}
}

func printReport(report *Report) {
	fmt.Println("\nTest Report")
	fmt.Println("-------------------")
	fmt.Printf("Total execution time: %v\n", report.totalTime)
	fmt.Printf("Total number of requests made: %d\n", report.totalRequests)
	fmt.Printf("Number of requests with HTTP status 200: %d\n", report.status200)
	fmt.Println("Distribution of other HTTP status codes:")
	for status, count := range report.statusOther {
		fmt.Printf("Status %d: %d requests\n", status, count)
	}
}
