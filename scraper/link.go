package scraper

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"github.com/gocolly/colly"
	"github.com/example/engines/model"
)

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
		log.Printf("[crawler1] visitando %s", r.URL.String())
	})

	c.OnHTML("div[class=searchColOne]", func(h *colly.HTMLElement) {
		h.ForEach("div", func(_ int, h *colly.HTMLElement) {
			text := h.ChildText("a")
			href := h.ChildAttr("a", "href")

			if text == "Engine" && href != "" {
				e := model.Engine{Name: text, URL: href}
				mu.Lock()
				links = append(links, e)
				mu.Unlock()
				c.Visit(h.Request.AbsoluteURL(href))
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("[crawler1] error en %s: %v", r.Request.URL, err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("[crawler1] respuesta de %s — status %d", r.Request.URL, r.StatusCode)
	})

	c.OnScraped(func(r *colly.Response) {
		log.Printf("[crawler1] terminado %s", r.Request.URL)
	})

	if err := c.PostMultipart(model.BaseURL+"/Home", map[string][]byte{
		"hdnVIN": []byte(vin),
	}); err != nil {
		return "", fmt.Errorf("error enviando VIN al formulario: %w", err)
	}

	c.Wait()

	if len(links) < 2 {
		if len(links) == 1 {
			return links[0].URL, nil
		}
		return "", errors.New("no se encontró ningún link de motores para ese VIN")
	}

	return links[1].URL, nil
}
