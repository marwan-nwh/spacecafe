package gravity

import (
	"crypto/md5"
	"fmt"
	"github.com/mmcdole/gofeed"
	"log"
	"time"
)

const (
	Workers       = 10
	SleepDuration = 30 * time.Minute
)

var feedParser = gofeed.NewParser()

func Parse(url string) (*gofeed.Feed, error) {
	feed, err := feedParser.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// return url hashed by md5, which is used as feed id
func MD5(url string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(url)))
}

type News chan *gofeed.Feed

var news = make(News)

// initiate workers pool
func Init() News {
	for w := 1; w <= Workers; w++ {
		worker()
	}
	return news
}

var jobs = make(chan *job)

type job struct {
	id        string
	url       string
	lastTitle string
}

// start new worker or a goroutine that receives messages
// on the feeds channel and update them
func worker() {
	go func() {
		for {
			select {
			case job := <-jobs:
				update(job)
			}
		}
	}()
}

var deads map[string]bool

func Kill(id string) {
	deads[id] = true
}

// start new goroutine for the feed
// that send job to the workers to update the feed,
// then sleep for a certain time, and repeat
func Pull(id string, url string, lastTitle ...string) {
	job := &job{id: id, url: url}
	if len(lastTitle) > 0 {
		job.lastTitle = lastTitle[0]
	}
	go func() {
		for {
			// log.Println("cheching " + url)
			if deads[job.id] {
				delete(deads, job.id)
				return
			}
			time.Sleep(SleepDuration)
			jobs <- job
		}
	}()

	// log.Println("pulling " + url)
}

func update(job *job) {
	// start := time.Now()
	feed, err := Parse(job.url)
	if err != nil {
		// log error
		return
	}
	if len(feed.Items) == 0 {
		return
	}
	if job.lastTitle == feed.Items[0].Title { // no updates
		return
	}

	log.Println("**********************************************")

	log.Println(job.url)
	for _, item := range feed.Items {
		log.Println(item.Title)
		log.Println(item.Published)
		log.Println(item.Updated)
	}

	// delete old items to send only new items
	for i, item := range feed.Items {
		if job.lastTitle == item.Title {
			feed.Items = feed.Items[:i-1]
			break
		}
	}

	log.Println("----------------------------------------------")

	job.lastTitle = feed.Items[0].Title

	// elapsed := time.Since(start)
	// fmt.Printf("%s Updated in %v sec\n", job.url, elapsed.Seconds())
	feed.Link = job.id // Todo: do something better than this cheap trick

	log.Println("New items")
	log.Println("last title:" + job.lastTitle)
	for _, item := range feed.Items {
		log.Println(item.Title)
		log.Println(item.Published)
		log.Println(item.Updated)
	}
	news <- feed
}
