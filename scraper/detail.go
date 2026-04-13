package scraper

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/EnriqueBravo115/crawler/model"
	"github.com/gocolly/colly"
)

// NOTE: 
// GetEngineDetails visits each engine page in parallel and extracts price, image, shipping and grade information.
func GetEngineDetails(links []model.Engine) ([]model.Engine, error) {
	var (
		mu      sync.Mutex
		engines []model.Engine
		wg      sync.WaitGroup
	)

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	c.SetRequestTimeout(model.RequestTimeout)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*hollanderparts.com*",
		Parallelism: 3,
		Delay:       500 * time.Millisecond,
	})

	// NOTE: OnHTML can fire 0, 1 or N times per page depending on how many
	// div[class=individualPartHolder] elements exist — this is NOT the right
	// place to call wg.Done().
	// We only collect data here.
	c.OnHTML("div[class=individualPartHolder]", func(h *colly.HTMLElement) {
		urlParts := strings.Split(h.Response.Request.URL.String(), "/")
		name := urlParts[len(urlParts)-1]

		e := model.Engine{
			Name:     name,
			URL:      h.Response.Request.URL.String(),
			Grade:    h.ChildText("div[class=gradeText]"),
			Img:      h.ChildAttr("img", "src"),
			Price:    h.ChildText("div[class=partPrice]"),
			Shipping: h.ChildText("div[class=partShipping]"),
		}

		mu.Lock()
		engines = append(engines, e)
		mu.Unlock()
	})

	// NOTE: OnScraped fires exactly ONCE per visited URL, after colly has finished
	// processing the entire page (including HTML).
	// This is the only safe place to call wg.Done().
	c.OnScraped(func(r *colly.Response) {
		log.Printf("[crawler3] finished %s", r.Request.URL)
		wg.Done()
	})

	// NOTE: OnError also fires once per failed URL instead of OnScraped,
	// so it needs its own Done() call.
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("[crawler3] error on %s: %v", r.Request.URL, err)
		wg.Done()
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("[crawler3] visiting %s", r.URL.String())
	})

	// NOTE: Visit all engine detail pages in parallel
	for _, engine := range links {
		if engine.URL == "" {
			continue
		}

		url := model.BaseURL + "/" + strings.TrimPrefix(engine.URL, "/")

		wg.Add(1)
		if err := c.Visit(url); err != nil {
			// NOTE: Visit failed before queuing the request (invalid URL, etc.)
			wg.Done()
			log.Printf("[crawler3] error queuing %s: %v", url, err)
		}
	}

	// NOTE: c.Wait() waits for colly's internal goroutines to finish.
	// wg.Wait() waits for OnScraped/OnError to call Done() for every URL.
	// Both are necessary because c.Wait() does not guarantee callbacks have finished.
	c.Wait()
	wg.Wait()

	if len(engines) == 0 {
		return nil, errors.New("no engine data was retrieved")
	}

	log.Printf("[crawler3] %d engines retrieved", len(engines))
	return engines, nil
}
