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

var visited = make([][]bool, 8)

var dx = 0

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	defer g.Close()

	for i := range visited {
		visited[i] = make([]bool, 8)
		for j := range visited[i] {
			visited[i][j] = false
		}
	}

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

	if err := g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, moveLeft); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, moveRight); err != nil {
		log.Panicln(err)
	}

	targetFps := int64(6)
	duration := time.Duration(time.Second.Nanoseconds()/targetFps) * time.Nanosecond
	tick := time.NewTicker(duration)
	last := time.Now()

	steps := 0
	f := uint8(7)
	b := uint8(7)

	go func() {
		for t := range tick.C {
			b = (b + 1) % 8

			if b == 0 {
				f = (f + 1) % 8
			}

			if b == f {
				b = (b + 1) % 8
				if b == 0 {
					f = 1
					b = 0
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
				fmt.Fprintln(v, dx, runtime.GOOS, "TARGET FPS:", targetFps, "FRAMELENGTH:", duration)
				fmt.Fprintln(v, "APPROX FPS:", fps)
				last = t
				return nil
			})

			if steps < 64 {
				steps += 1
				g.Execute(func(g *gocui.Gui) error {
					v, err := g.View("grid")
					if err != nil {
						log.Panicln(err)
					}
					v.Clear()
					for i := range visited {
						for j := range visited[i] {
							if visited[i][j] {
								fmt.Fprint(v, colorize("▮", uint8(j), uint8(i)))
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
	if v, err := g.SetView("t1", maxX/2-20+dx, maxY/2, maxX/2+30+dx, maxY/2+3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
	}
	if v, err := g.SetView("t2", maxX/2-20+dx, 0, maxX/2+20+dx, maxY/2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Frame = true
		v.Title = "]TIME["
	}
	if v, err := g.SetView("grid", 0, 0+dx, 9, 9+dx); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.Title = "]SEEN["
	}
	return nil
}

func moveLeft(_ *gocui.Gui, _ *gocui.View) error {
	dx -= 1
	return nil
}

func moveRight(_ *gocui.Gui, _ *gocui.View) error {
	dx += 1
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
