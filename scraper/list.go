package scraper

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/EnriqueBravo115/crawler/model"
)

// NOTE: GetEngineList visits the engines listing page and returns
// all found engines with their individual detail URLs.
func GetEngineList(urlSuffix string) ([]model.Engine, error) {
	var (
		mu      sync.Mutex
		engines []model.Engine
	)

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	c.SetRequestTimeout(model.RequestTimeout)

	// NOTE: Extract all engine links from the listing page
	c.OnHTML("div[class=searchColOne]", func(h *colly.HTMLElement) {
		h.ForEach("div", func(_ int, h *colly.HTMLElement) {
			name := h.ChildText("a")
			href := h.ChildAttr("a", "href")

			if name == "" || href == "" {
				return
			}

			e := model.Engine{Name: name, URL: href}

			mu.Lock()
			engines = append(engines, e)
			mu.Unlock()

			// NOTE: Visit the individual engine page
			c.Visit(h.Request.AbsoluteURL(href))
		})
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("[crawler2] visiting %s", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("[crawler2] error on %s: %v", r.Request.URL, err)
	})

	// NOTE: Build the full URL
	url := model.BaseURL + "/" + strings.TrimPrefix(urlSuffix, "/")

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("error visiting engines listing page: %w", err)
	}

	c.Wait()

	return engines, nil
}
