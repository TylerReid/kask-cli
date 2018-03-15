package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
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
		if v, err := g.SetView(string(k.KegId), maxX/4, 0, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = k.Tap.Description
			v.Autoscroll = true
			v.Editable = false
			v.Wrap = false
		}
	}

	if v, err := g.SetView("TapList", 0, 0, maxX/4-1, maxY-1); err != nil {
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

	v, err := g.View("TapList")
	if err != nil {
		return err
	}
	v.Clear()
	for i, k := range kegs {
		if i == currentView {
			//use color code to set text highlight for selected command
			fmt.Fprintf(v, "\033[32;7m%v\033[0m\n", k.Beer.BeerName)
		} else {
			fmt.Fprintln(v, k.Beer.BeerName)
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
