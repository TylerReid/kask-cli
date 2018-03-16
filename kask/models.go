package kask

import (
	"image"
)

type Tap struct {
	TapId       int
	TapName     string
	Description string
}

type KegOnTap struct {
	Active   int
	NetVote  int
	UserVote int
	Keg      struct {
		KegId         int
		Size          string
		InitialVolume float64
		RemovedVolume float64
		TapId         int
		Beer          Beer
		Tap           Tap
	}
}

type Beer struct {
	BeerId          int
	BeerName        string
	BeerDescription string
	ABV             float64
	LabelUrl        string
	Brewery         Brewery
	Style           Style
	ImageData       image.Image
}

type Brewery struct {
	BreweryId          int
	BreweryName        string
	BreweryDescription string
	Image              string
	Website            string
	ImageData          image.Image
}

type Style struct {
	StyleId          int
	StyleName        string
	StyleDescription string
}
