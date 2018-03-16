package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/TylerReid/kask-cli/kask"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
	"github.com/zyxar/image2ascii/ascii"
)

var kaskApi kask.Kask
var kegs []kask.KegOnTap

func main() {
	kaskUrl := *flag.String("kaskurl", "https://kask.kabbage.com/api", "kask api url")

	kaskApi = kask.Kask{Url: kaskUrl}

	var err error
	kegs, err = kaskApi.GetBeersOnTap()
	if err != nil || len(kegs) == 0 {
		fmt.Println("No Kegs! ☹️")
		return
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.BgColor = gocui.ColorBlack
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
		//Description
		if v, err := g.SetView(viewKey(k), maxX/6, 0, maxX-1, maxY-int((float64(maxY)*0.7))-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = k.Keg.Tap.Description
			v.Autoscroll = false
			v.Editable = false
			v.Wrap = true
			//set FgColor to bold to let us handle colors with color codes. Should look into lib PR to make this not dumb
			v.FgColor = gocui.AttrBold
			populateKegInfo(v, k)
		}
		//Fill meter
		if v, err := g.SetView(volumeKey(k), 0, maxY-int((float64(maxY)*0.1)), maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Fill Level"
			v.Autoscroll = false
			v.Editable = false
			v.Wrap = false
			v.FgColor = gocui.ColorGreen
			populateVolumeInfo(v, k)
		}
		//Image
		v, err := g.SetView(imageKey(k), maxX/6, maxY-int((float64(maxY)*0.7)), maxX-1, maxY-int((float64(maxY)*0.1))-1)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Autoscroll = false
			v.Editable = false
			v.Wrap = false
			v.FgColor = gocui.AttrBold
		}
		populateImage(v, k)
	}
	//Tap List
	if v, err := g.SetView("TapList", 0, 0, maxX/6-1, maxY-int((float64(maxY)*0.1))-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Taps"
		v.Autoscroll = true
		v.Editable = false
		v.Wrap = true
		v.FgColor = gocui.AttrBold
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
	_, err := g.SetCurrentView(viewKey(kegs[currentView]))
	if err != nil {
		return err
	}
	_, err = g.SetViewOnTop(viewKey(kegs[currentView]))
	_, err = g.SetViewOnTop(volumeKey(kegs[currentView]))
	_, err = g.SetViewOnTop(imageKey(kegs[currentView]))

	v, err := g.View("TapList")
	if err != nil {
		return err
	}
	v.Clear()
	for i, k := range kegs {
		if i == currentView {
			color.New(color.BgGreen, color.FgWhite).Fprintf(v, "%v\n", k.Keg.Beer.BeerName)
		} else {
			fmt.Fprintln(v, k.Keg.Beer.BeerName)
		}
	}
	return nil
}

func populateKegInfo(v *gocui.View, k kask.KegOnTap) {
	x, _ := v.Size()
	dividerString := strings.Repeat("~", x)
	v.Clear()
	color.New(color.FgRed).Fprintf(v, "%v\n", dividerString)
	fmt.Fprintf(v, "%v\n", k.Keg.Beer.BeerName)
	fmt.Fprintf(v, "%v\n\n", k.Keg.Beer.Brewery.BreweryName)
	ratingSign := ""
	if k.NetVote > 0 {
		ratingSign = "+"
	}
	fmt.Fprintf(v, "%v Barrel %v ABV %v%v Rating\n\n", k.Keg.Size, k.Keg.Beer.ABV, ratingSign, k.NetVote)
	color.New(color.FgRed).Fprintf(v, "%v\n", dividerString)
	fmt.Fprintf(v, "%v\n\n", k.Keg.Beer.BeerDescription)
	color.New(color.FgCyan).Fprintf(v, "%v\n\n", k.Keg.Beer.Brewery.Website)
}

func populateVolumeInfo(v *gocui.View, k kask.KegOnTap) {
	v.Clear()
	maxWidth, _ := v.Size()
	percentLeft := k.Keg.RemovedVolume / k.Keg.InitialVolume
	fillLevel := maxWidth - int(float64(maxWidth)*percentLeft)
	for i := 0; i < maxWidth; i++ {
		if i < fillLevel {
			fmt.Fprint(v, "▓")
		} else {
			fmt.Fprint(v, "░")
		}
	}
}

func populateImage(v *gocui.View, k kask.KegOnTap) {
	v.Clear()
	x, y := v.Size()
	if k.Keg.Beer.Brewery.ImageData != nil {
		ascii.Encode(v, k.Keg.Beer.Brewery.ImageData, ascii.Options{Width: x, Height: y})
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func viewKey(k kask.KegOnTap) string {
	return string(k.Keg.KegId)
}

func volumeKey(k kask.KegOnTap) string {
	return string(k.Keg.KegId) + "volume"
}

func imageKey(k kask.KegOnTap) string {
	return string(k.Keg.KegId) + "image"
}
