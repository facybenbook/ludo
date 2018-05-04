package main

import (
	"fmt"
	"io/ioutil"
	"libretro"

	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

type menuCallback func()

type entry struct {
	label       string
	scroll      float32
	scrollTween *gween.Tween
	ptr         int
	callback    menuCallback
	children    []entry
}

var menuStack []entry

func buildExplorer(path string) entry {
	var menu entry
	menu.label = "Explorer"

	files, err := ioutil.ReadDir(path)
	if err != nil {
		notify(err.Error(), 120)
		fmt.Println(err)
	}

	for _, f := range files {
		f := f
		menu.children = append(menu.children, entry{
			label: f.Name(),
			callback: func() {
				if f.IsDir() {
					menuStack = append(menuStack, buildExplorer(path+"/"+f.Name()+"/"))
				}
			},
		})
	}

	return menu
}

func buildMainMenu() entry {
	var menu entry
	menu.label = "Main Menu"

	if g.coreRunning {
		menu.children = append(menu.children, entry{
			label: "Quick Menu",
			callback: func() {
				menuStack = append(menuStack, buildQuickMenu())
			},
		})
	}

	menu.children = append(menu.children, entry{
		label: "Load Core",
		callback: func() {
			menuStack = append(menuStack, buildExplorer("./cores"))
		},
	})

	menu.children = append(menu.children, entry{
		label: "Load Game",
		callback: func() {
			menuStack = append(menuStack, buildExplorer("./roms"))
		},
	})

	menu.children = append(menu.children, entry{
		label: "Help",
		callback: func() {
			notify("Not implemented yet", 120)
		},
	})

	menu.children = append(menu.children, entry{
		label: "Quit",
		callback: func() {
			window.SetShouldClose(true)
		},
	})

	return menu
}

func buildQuickMenu() entry {
	var menu entry
	menu.label = "Quick Menu"

	menu.children = append(menu.children, entry{
		label: "Resume",
		callback: func() {
			g.menuActive = !g.menuActive
		},
	})

	menu.children = append(menu.children, entry{
		label: "Save State",
		callback: func() {
			fmt.Println("[Menu]: Not implemented")
			notify("Not implemented", 120)
		},
	})

	menu.children = append(menu.children, entry{
		label: "Load State",
		callback: func() {
			fmt.Println("[Menu]: Not implemented")
			notify("Not implemented", 120)
		},
	})

	menu.children = append(menu.children, entry{
		label: "Take Screenshot",
		callback: func() {
			fmt.Println("[Menu]: Not implemented")
			notify("Not implemented", 120)
		},
	})

	return menu
}

var vSpacing = 70
var inputCooldown = 0

func menuInput() {
	currentMenu := &menuStack[len(menuStack)-1]

	if inputCooldown > 0 {
		inputCooldown--
	}

	if newState[0][libretro.DeviceIDJoypadDown] && inputCooldown == 0 {
		currentMenu.ptr++
		if currentMenu.ptr >= len(currentMenu.children) {
			currentMenu.ptr = 0
		}
		currentMenu.scrollTween = gween.New(currentMenu.scroll, float32(currentMenu.ptr*vSpacing), 0.15, ease.OutSine)
		inputCooldown = 10
	}

	if newState[0][libretro.DeviceIDJoypadUp] && inputCooldown == 0 {
		currentMenu.ptr--
		if currentMenu.ptr < 0 {
			currentMenu.ptr = len(currentMenu.children) - 1
		}
		currentMenu.scrollTween = gween.New(currentMenu.scroll, float32(currentMenu.ptr*vSpacing), 0.10, ease.OutSine)
		inputCooldown = 10
	}

	if released[0][libretro.DeviceIDJoypadA] {
		if currentMenu.children[currentMenu.ptr].callback != nil {
			currentMenu.children[currentMenu.ptr].callback()
		}
	}

	if released[0][libretro.DeviceIDJoypadB] {
		if len(menuStack) > 1 {
			menuStack = menuStack[:len(menuStack)-1]
		}
	}
}

func renderMenuList() {
	_, h := window.GetFramebufferSize()
	fullscreenViewport()

	currentMenu := &menuStack[len(menuStack)-1]
	if currentMenu.scrollTween != nil {
		currentMenu.scroll, _ = currentMenu.scrollTween.Update(1.0 / 60.0)
	}

	video.font.SetColor(1, 1, 1, 1.0)
	video.font.Printf(60, 20+60, 0.5, currentMenu.label)

	for i, e := range currentMenu.children {
		y := -currentMenu.scroll + 20 + float32(vSpacing*(i+2))

		if y < 0 || y > float32(h) {
			continue
		}

		if i == currentMenu.ptr {
			video.font.SetColor(0.0, 1.0, 0.0, 1.0)
		} else {
			video.font.SetColor(0.6, 0.6, 0.9, 1.0)
		}
		video.font.Printf(100, y, 0.5, e.label)
	}
}

func menuInit() {
	if g.coreRunning {
		menuStack = append(menuStack, buildMainMenu())
		menuStack = append(menuStack, buildQuickMenu())
	} else {
		menuStack = append(menuStack, buildMainMenu())
	}
}