package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/zyxar/image2ascii/ascii"
)

var kaskUrl string
var taps []tap
var kegs []keg

func main() {
	kaskUrl = *flag.String("kaskurl", "https://kask.kabbage.com/api", "kask api url")

	taps = getTaps(kaskUrl)
	if len(taps) == 0 {
		log.Panicln("No Taps! ☹️")
	}

	kegs = getKegs(kaskUrl, taps)
	if len(kegs) == 0 {
		log.Panicln("No Kegs! ☹️")
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, switchWindowDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, switchWindowUp); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	for _, k := range kegs {
		if v, err := g.SetView(string(k.KegId), maxX/6, 0, maxX-1, maxY-int((float64(maxY)*0.7))-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = k.Tap.Description
			v.Autoscroll = true
			v.Editable = false
			v.Wrap = true
			populateKegInfo(v, k)
		}

		if v, err := g.SetView(string(k.KegId)+"volume", maxX/6, maxY-int((float64(maxY)*0.1)), maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Fill Level"
			v.Autoscroll = true
			v.Editable = false
			v.Wrap = false
			v.FgColor = gocui.ColorGreen
			populateVolumeInfo(v, k)
		}

		if v, err := g.SetView(string(k.KegId)+"image", maxX/6, maxY-int((float64(maxY)*0.7)), maxX-1, maxY-int((float64(maxY)*0.1))-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Autoscroll = true
			v.Editable = false
			v.Wrap = false
			v.FgColor = gocui.ColorWhite
			populateImage(v, k)
		}
	}

	if v, err := g.SetView("TapList", 0, 0, maxX/6-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Taps"
		v.Autoscroll = true
		v.Editable = false
		v.Wrap = true
	}

	if err := setCurrentWindow(g); err != nil {
		return err
	}

	return nil
}

var currentView = 0

func switchWindowDown(g *gocui.Gui, v *gocui.View) error {
	currentView++
	if currentView >= len(kegs) {
		currentView = 0
	}
	if err := setCurrentWindow(g); err != nil {
		return err
	}
	return nil
}

func switchWindowUp(g *gocui.Gui, v *gocui.View) error {
	currentView--
	if currentView < 0 {
		currentView = len(kegs) - 1
	}
	if err := setCurrentWindow(g); err != nil {
		return err
	}
	return nil
}

func setCurrentWindow(g *gocui.Gui) error {
	_, err := g.SetCurrentView(string(kegs[currentView].KegId))
	if err != nil {
		return err
	}
	_, err = g.SetViewOnTop(string(kegs[currentView].KegId))
	_, err = g.SetViewOnTop(string(kegs[currentView].KegId) + "volume")
	_, err = g.SetViewOnTop(string(kegs[currentView].KegId) + "image")

	v, err := g.View("TapList")
	if err != nil {
		return err
	}
	v.Clear()
	for i, k := range kegs {
		if i == currentView {
			color.New(color.BgGreen, color.FgHiWhite).Fprintf(v, "%v\n", k.Beer.BeerName)
		} else {
			fmt.Fprintln(v, k.Beer.BeerName)
		}
	}
	return nil
}

func populateKegInfo(v *gocui.View, k keg) {
	v.Clear()
	fmt.Println()
	fmt.Fprintf(v, "%v\n", k.Beer.BeerName)
	fmt.Fprintf(v, "%v\n\n", k.Beer.Brewery.BreweryName)
	fmt.Fprintf(v, "%v Barrel\n\n", k.Size)
	fmt.Fprint(v, "~~~~~~~~~~~~~~~~*~~~~~~~~~~~~~~~~\n\n")
	fmt.Fprintf(v, "%v\n\n", k.Beer.BeerDescription)
	fmt.Fprintf(v, "%v\n\n", k.Beer.Brewery.Website)
}

func populateVolumeInfo(v *gocui.View, k keg) {
	v.Clear()
	maxWidth, _ := v.Size()
	percentLeft := k.RemovedVolume / k.InitialVolume
	fillLevel := maxWidth - int(float64(maxWidth)*percentLeft)
	for i := 0; i < maxWidth; i++ {
		if i < fillLevel {
			fmt.Fprint(v, "▓")
		} else {
			fmt.Fprint(v, "░")
		}
	}
}

func populateImage(v *gocui.View, k keg) {
	v.Clear()
	x, y := v.Size()
	if k.Beer.Brewery.ImageData != nil {
		ascii.Encode(v, k.Beer.Brewery.ImageData, ascii.Options{Width: x, Height: y})
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
