package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Page struct {
	URL  string
	Body string
}

func main() {
	pageChan := make(chan Page, 1000)
	// resultChan := make(chan string, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	counter := &VisitCounter{}
	// Start the monitoring goroutine
	go monitorVisits(ctx, counter, 5*time.Second)

	var parserWg sync.WaitGroup

	crawler := NewCrawler(pageChan, counter)
	crawler.Crawl(ctx)

	// Start parser goroutines
	// for i := 0; i < 5; i++ {
	// 	parserWg.Add(1)
	// 	go parser(pageChan, resultChan, &parserWg)
	// }

	go func() {
		<-ctx.Done()
		parserWg.Wait()
		close(pageChan)
		fmt.Println("Crawling timed out")
	}()

	// Print results
	// for result := range resultChan {
	// 	fmt.Println(result)
	// }
	// close(resultChan)
}

// func parser(pageChan <-chan Page, resultChan chan<- string, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for page := range pageChan {
// 		// Simulate parsing work
// 		resultChan <- fmt.Sprintf("Parsed: %s", page.URL)
// 	}
// }
