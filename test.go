package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jroimartin/gocui"
)

var tFmt = "2006-01-02 15:04:05.000000000 -0700 MST"

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	t1 := time.NewTicker(time.Millisecond * 100)
	t2 := time.NewTicker(time.Millisecond * 350)

	go func() {
		for t := range t1.C {
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("t1")
				if err != nil {
					log.Panicln(err)
				}
				v.Clear()
				fmt.Fprintln(v, t.Format(tFmt))
				return nil
			})
		}
	}()

	go func() {
		for t := range t2.C {
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("t2")
				if err != nil {
					log.Panicln(err)
				}
				v.Clear()
				fmt.Fprintln(v, t.Format(tFmt))
				return nil
			})
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("t1", maxX/2-20, maxY/2, maxX/2+20, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, time.Now().Format(tFmt))
	}
	if v, err := g.SetView("t2", maxX/2-20, maxY/4, maxX/2+20, maxY/4+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, time.Now().Format(tFmt))
	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
