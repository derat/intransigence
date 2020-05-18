// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

type structData struct {
	Context          string `json:"@context"`
	Type             string `json:"@type"`
	MainEntityOfPage string `json:"mainEntityOfPage"`
	Headline         string `json:"headline"`
	Description      string `json:"description,omitempty"`
	DateModified     string `json:"dateModified,omitempty"`
	DatePublished    string `json:"datePublished"`

	Author    structDataAuthor    `json:"author"`
	Publisher structDataPublisher `json:"publisher"`

	Image *structDataImage `json:"image,omitempty"`
}

type structDataAuthor struct {
	Type  string `json:"@type"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type structDataPublisher struct {
	Type string          `json:"@type"`
	Name string          `json:"name"`
	URL  string          `json:"url"`
	Logo structDataImage `json:"logo"`
}

type structDataImage struct {
	Type   string `json:"@type"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
