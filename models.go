package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type tap struct {
	TapId       int
	TapName     string
	Description string
}

type kegResponse struct {
	Active   int
	NetVote  int
	UserVote int
	Keg      keg
}

type keg struct {
	KegId         int
	Size          string
	InitialVolume float64
	RemovedVolume float64
	TapId         int
	Beer          beer
	Tap           tap
}

type beer struct {
	BeerId          int
	BeerName        string
	BeerDescription string
	ABV             float64
	LabelUrl        string
	Brewery         brewery
	Style           style
}

type brewery struct {
	BreweryId          int
	BreweryName        string
	BreweryDescription string
	Image              string
	Website            string
}

type style struct {
	StyleId          int
	StyleName        string
	StyleDescription string
}

func getTaps(baseUrl string) []tap {
	url := fmt.Sprintf("%s/beers/taps", baseUrl)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	var taps []tap
	json.NewDecoder(res.Body).Decode(&taps)
	return taps
}

func getKegs(baseUrl string, taps []tap) []keg {
	var kegs []keg
	for _, t := range taps {
		url := fmt.Sprintf("%s/beers/contents/tap/%v", baseUrl, t.TapId)
		res, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
		var k kegResponse
		json.NewDecoder(res.Body).Decode(&k)
		k.Keg.Tap = t
		kegs = append(kegs, k.Keg)
	}
	return kegs
}
