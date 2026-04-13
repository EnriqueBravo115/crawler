package model

import "time"

type Engine struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Price    string `json:"price,omitempty"`
	Shipping string `json:"shipping,omitempty"`
	Img      string `json:"img,omitempty"`
	Grade    string `json:"grade,omitempty"`
}

const (
	BaseURL        = "https://www.hollanderparts.com"
	RequestTimeout = 15 * time.Second
)
