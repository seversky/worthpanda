package scrape

import (
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/seversky/gachifinder"
)

var _ gachifinder.Scraper = &Scrape{}

// Scrape struct.
type Scrape struct {
	VisitDomains	[]string
	AllowedDomains	[]string

	// Unexported ...
	c 			*colly.Collector	// Will be assigned by inside Do func.
	timestamp 	string
}

// Do creates colly.collector and queue, and then do and wait till done
func (s *Scrape) Do(f gachifinder.ParsingHandler) (<-chan gachifinder.GachiData) {
	cd := make(chan gachifinder.GachiData)

	go func () {
		// Record the beginning time.
		s.timestamp = time.Now().UTC().Format("2006-01-02T15:04:05")
		fmt.Println("I! It gets begun at", time.Now())

		// Instantiate default collector
		s.c = colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"),
			colly.Async(true),
			colly.MaxDepth(1),
			colly.AllowedDomains(s.AllowedDomains...),
		)

		s.c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 1,
			Delay: 100 * time.Millisecond,
			RandomDelay: 2 * time.Second,
		})

		// create a request queue with 1 consumer threads
		q, err := queue.New(
			1, // Number of consumer threads
			&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
		)
		if err != nil {
			fmt.Println("Creating Queue is Failed:", err)
			panic(err)
		}

		for _, url := range s.VisitDomains {
			err := q.AddURL(url)
			if err != nil {
				fmt.Println("Adding url into the queue is Failed:", err)
				panic(err)
			}
		}

		f(cd)

		// Consume URLs.
		err = q.Run(s.c)
		if err != nil {
			fmt.Println("Running the queue is Failed:", err)
			panic(err)
		}
		// Wait for the crawling to complete.
		s.c.Wait()

		close(cd)
	}()

	return cd
}

// ParsingHandler is an abstract function.
// this has to be implemented into the embedded(is-a) method.
func (s *Scrape) ParsingHandler(chan<- gachifinder.GachiData) {}