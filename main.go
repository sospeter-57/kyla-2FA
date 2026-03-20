package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func main() {

	// set up and initialize application and main window
	app := app.New()
	window := app.NewWindow("kyla-2FA")
	window.Resize(fyne.NewSize(600, 400))
	window.CenterOnScreen()

	// fetch application logo
	logo := getLogo()

	// create application's main container for main window
	container := container.NewVBox()
	container.Add(&logo)

	// set contents to main window, show and run window
	window.SetContent(container)
	window.Show()
	app.Run()
}

func getLogo() canvas.Image {
	logo := canvas.NewImageFromFile("assets/kyla-2FA_logo.png")
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(200, 100))
	return *logo
}
