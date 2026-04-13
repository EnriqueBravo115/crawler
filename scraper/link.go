package scraper

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gocolly/colly"
	"github.com/EnriqueBravo115/crawler/model"
)

// NOTE: GetEngineLink submits the VIN to the search form and returns the URL
// of the engines page (the second link that appears after the search).
func GetEngineLink(vin string) (string, error) {
	var (
		mu    sync.Mutex
		links []model.Engine
	)

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	c.SetRequestTimeout(model.RequestTimeout)

	c.OnRequest(func(r *colly.Request) {
		log.Printf("[crawler1] visiting %s", r.URL.String())
	})

	// NOTE: Look for the "Engine" link in the search results
	c.OnHTML("div[class=searchColOne]", func(h *colly.HTMLElement) {
		h.ForEach("div", func(_ int, h *colly.HTMLElement) {
			text := h.ChildText("a")
			href := h.ChildAttr("a", "href")

			if text == "Engine" && href != "" {
				e := model.Engine{Name: text, URL: href}

				mu.Lock()
				links = append(links, e)
				mu.Unlock()

				// NOTE: Visit the engines page immediately when we find the link
				c.Visit(h.Request.AbsoluteURL(href))
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("[crawler1] error on %s: %v", r.Request.URL, err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("[crawler1] response from %s — status %d", r.Request.URL, r.StatusCode)
	})

	c.OnScraped(func(r *colly.Response) {
		log.Printf("[crawler1] finished %s", r.Request.URL)
	})

	// NOTE: Submit the VIN using POST multipart form
	if err := c.PostMultipart(model.BaseURL+"/Home", map[string][]byte{
		"hdnVIN": []byte(vin),
	}); err != nil {
		return "", fmt.Errorf("error submitting VIN to the form: %w", err)
	}

	c.Wait()

	// NOTE: 
	// If we only found one link, return it.
	// Otherwise return the second link (usually the "Engine" one).
	if len(links) < 2 {
		if len(links) == 1 {
			return links[0].URL, nil
		}
		return "", errors.New("no engine link found for this VIN")
	}

	return links[1].URL, nil
}
