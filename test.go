package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jroimartin/gocui"
	"runtime"
)

//                                     ,-windows precision
//                                    |  ,-mac precision
//                                    v v
var tFmt = "2006-01-02 15:04:05.000000000 -0700 MST"

var dx = 0
var dy = 0

var pause bool = false

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	visited := initVisited()

	g.SetManagerFunc(layout)

	if runtime.GOOS == "windows" {
		if err := g.SetKeybinding("", gocui.KeyCtrlW, gocui.ModNone, quit); err != nil {
			log.Panicln(err)
		}
	} else {
		if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
			log.Panicln(err)
		}
	}

	bindMoveKey := func(k gocui.Key, f func()) {
		ew := func(_ *gocui.Gui, _ *gocui.View) error { f(); return nil }
		if err := g.SetKeybinding("", k, gocui.ModNone, ew); err != nil {
			log.Panicln(err)
		}
	}

	bindMoveKey(gocui.KeyArrowLeft, func() { dx -= 1 })
	bindMoveKey(gocui.KeyArrowRight, func() { dx += 1 })
	bindMoveKey(gocui.KeyArrowUp, func() { dy -= 1 })
	bindMoveKey(gocui.KeyArrowDown, func() { dy += 1 })
	bindMoveKey(gocui.KeySpace, func() { pause = !pause})

	targetFps := int64(6)
	duration := time.Duration(time.Second.Nanoseconds()/targetFps) * time.Nanosecond
	tick := time.NewTicker(duration)
	last := time.Now()

	f := uint8(7)
	b := uint8(7)

	step := 0

	go func() {
		for t := range tick.C {
			step += 1
			if !pause {
				b = (b + 1) % 8

				if b == 0 {
					f = (f + 1) % 8
				}

				if b == f {
					b = (b + 1) % 8
					if b == 0 {
						visited = initVisited()
						f = 0
						b = 1
					}
				}

				visited[b][f] = true

				g.Execute(func(g *gocui.Gui) error {
					v, err := g.View("t2")
					if err != nil {
						log.Panicln(err)
					}
					fmt.Fprint(v, "\n", colorize(t.Format(tFmt), b, f))
					return nil
				})

				g.Execute(func(g *gocui.Gui) error {
					v, err := g.View("t1")
					if err != nil {
						log.Panicln(err)
					}
					v.Clear()

					denominator := t.Sub(last).Nanoseconds()
					var fps interface{} = "INFINITY"
					if denominator != 0 {
						fps = time.Second.Nanoseconds() / denominator
					}
					fmt.Fprintln(v, runtime.GOOS, "TICK FPS:", fps, "dx:", dx, "dy:", dy)
					last = t
					return nil
				})

				g.Execute(func(g *gocui.Gui) error {
					v, err := g.View("grid")
					if err != nil {
						log.Panicln(err)
					}
					v.Clear()
					for i := range visited {
						for j := range visited[i] {
							if visited[i][j] {
								fmt.Fprint(v, colorize("▮", uint8(i), uint8(j)))
							} else {
								fmt.Fprint(v, " ")
							}
						}
						fmt.Fprint(v, "\n")
					}
					return nil
				})
			}
		}
	}()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func colorize(s string, f, b uint8) string {
	return fmt.Sprintf("\033[3%dm\033[4%dm%s\033[0m", f, b, s)
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("grid", 0, 0, 9, 9); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = "]SEEN["
	}
	if v, err := g.SetView("t1", 0, maxY-2, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}
	if v, err := g.SetView("t2", maxX/2-20+dx, 0+dy, maxX/2+20+dx, maxY/2+dy); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Frame = true
		v.Title = "]TIME["
	}
	return nil
}

func initVisited() [][]bool {
	var visited = make([][]bool, 8)
	for i := range visited {
		visited[i] = make([]bool, 8)
		for j := range visited[i] {
			visited[i][j] = false
		}
	}
	return visited
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
