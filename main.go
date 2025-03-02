package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL     = "https://openintel.nl/download/forward-dns/basis=toplist/source=%s/year=%d/month=%02d/day=%02d/"
	downloadDir = "parquet_files"
	defaultYear = 2016
	maxYear     = 2025
)

var datasets = []string{"alexa", "radar", "tranco", "umbrella"}
var workerLimit = 10 // Maximum number of concurrent downloads

// Global HTTP client
var httpClient *http.Client

func main() {
	// Define command-line arguments
	startYear := flag.Int("start-year", defaultYear, "Start year (minimum 2016)")
	endYear := flag.Int("end-year", maxYear, "End year (maximum 2025)")
	proxyURL := flag.String("proxy", "", "HTTP proxy URL (optional)")
	showHelp := flag.Bool("help", false, "Display help menu")

	flag.Parse()

	// Display help and exit if --help is passed
	if *showHelp {
		showUsage()
		return
	}

	// Validate the year range
	if *startYear < defaultYear || *endYear > maxYear || *startYear > *endYear {
		fmt.Println("❌ Error: Year range must be between 2016 and 2025.")
		showUsage()
		return
	}

	// Configure proxy if provided
	proxyFunc := http.ProxyFromEnvironment
	if *proxyURL != "" {
		proxy, err := url.Parse(*proxyURL)
		if err == nil {
			proxyFunc = http.ProxyURL(proxy)
			fmt.Println("🛡️ Using proxy:", *proxyURL)
		} else {
			fmt.Println("❌ Error configuring proxy:", err)
			return
		}
	}

	// Create HTTP client with proxy support
	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy:           proxyFunc,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip SSL certificate errors if needed
		},
		Timeout: 30 * time.Second, // Timeout to avoid blocking requests
	}

	// Create the download directory if it does not exist
	os.MkdirAll(downloadDir, os.ModePerm)

	// Display download info
	fmt.Println("📂 Download directory:", downloadDir)
	fmt.Printf("📅 Downloading files from %d to %d\n", *startYear, *endYear)

	// Concurrency control channel
	sem := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup

	// Loop through years, months, and days
	for year := *startYear; year <= *endYear; year++ {
		for month := 1; month <= 12; month++ {
			for day := 1; day <= 31; day++ {
				for _, dataset := range datasets {
					url := fmt.Sprintf(baseURL, dataset, year, month, day)

					// Add a worker goroutine
					wg.Add(1)
					sem <- struct{}{} // Limit concurrency

					go func(url string) {
						defer wg.Done()
						defer func() { <-sem }() // Free slot
						processPage(url)
					}(url)
				}
			}
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("✅ Process completed!")
}

// showUsage displays the help menu
func showUsage() {
	fmt.Println(`
Usage:
  programa [options]

Options:
  --start-year=N    Define the start year (minimum 2016)
  --end-year=N      Define the end year (maximum 2025)
  --proxy=URL       Use an HTTP proxy (optional)
  --help            Show this help menu

Example:
  programa --start-year=2020 --end-year=2022 --proxy=http://127.0.0.1:8080
`)
}

// processPage fetches the webpage and extracts .parquet file links
func processPage(url string) {
	fmt.Println("🌐 Checking:", url)

	// Create request with required cookie
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("❌ Error creating request:", url)
		return
	}
	req.Header.Set("Cookie", "openintel-data-agreement-accepted=true")

	// Execute HTTP request
	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("❌ Error accessing:", url)
		return
	}
	defer resp.Body.Close()

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("❌ Error processing HTML:", url)
		return
	}

	// Find links inside "flex-container" class
	doc.Find("a.flex-container").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			downloadFile(link)
		}
	})
}

// downloadFile downloads a file
func downloadFile(fileURL string) {
	fileName := filepath.Join(downloadDir, filepath.Base(fileURL))

	// Check if the file already exists
	if _, err := os.Stat(fileName); err == nil {
		fmt.Println("✅ File already downloaded:", fileName)
		return
	}

	fmt.Println("⬇️  Downloading:", fileURL)

	// Execute file download
	resp, err := http.Get(fileURL)
	if err != nil {
		fmt.Println("❌ Error downloading:", fileURL)
		return
	}
	defer resp.Body.Close()

	// Save the file to disk
	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println("❌ Error creating file:", fileName)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("❌ Error saving file:", fileName)
		return
	}

	fmt.Println("✅ Download completed:", fileName)
}
