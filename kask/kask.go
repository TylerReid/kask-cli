package kask

import (
	"encoding/json"
	"fmt"
	"image"
	"net/http"
)

type Kask struct {
	Url string
}

func (k *Kask) GetBeersOnTap() ([]KegOnTap, error) {
	taps, err := getTaps(k.Url)
	if err != nil {
		return nil, err
	}
	kegs, err := getKegs(k.Url, taps)
	if err != nil {
		return nil, err
	}
	return kegs, nil
}

func getTaps(baseUrl string) ([]Tap, error) {
	url := fmt.Sprintf("%s/beers/taps", baseUrl)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var taps []Tap
	json.NewDecoder(res.Body).Decode(&taps)
	return taps, nil
}

func getKegs(baseUrl string, taps []Tap) ([]KegOnTap, error) {
	var kegs []KegOnTap
	for _, t := range taps {
		url := fmt.Sprintf("%s/beers/contents/tap/%v", baseUrl, t.TapId)
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		var k KegOnTap
		json.NewDecoder(res.Body).Decode(&k)
		if k.Active == 0 {
			continue
		}
		k.Keg.Tap = t
		k.Keg.Beer.getBeerImage()
		k.Keg.Beer.Brewery.getKegImage()
		kegs = append(kegs, k)
	}
	return kegs, nil
}

func (b *Brewery) getKegImage() {
	b.ImageData = getImage(b.Image)
}

func (b *Beer) getBeerImage() {
	b.ImageData = getImage(b.LabelUrl)
}

func getImage(url string) image.Image {
	res, err := http.Get(url)
	if err != nil {
		//could be any random problem, just give up
		return nil
	}
	defer res.Body.Close()
	i, _, err := image.Decode(res.Body)
	if err != nil {
		return nil
	}
	return i
}
