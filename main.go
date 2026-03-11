package main

import (
	"fmt"
	"log"
	"time"

	"kyla-2FA/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/pquerna/otp/totp"
)

func main() {
	randomBytes := utils.GenerateRandomBytes(32)
	encodedSecret := utils.EncodeBase64(randomBytes)

	// create the main window
	app := app.New()
	window := app.NewWindow("kyla-2FA")
	window.CenterOnScreen()
	window.Resize(fyne.NewSize(700, 500))

	// create a label widget
	label := widget.NewLabel("just kidding here and there with fyne qt bindings")
	window.SetContent(label)

	code, err := totp.GenerateCode(encodedSecret, time.Now())
	if err != nil {
		log.Fatal("Error: ", err)
	}

	fmt.Println("TOTP code: ", code)
	window.ShowAndRun()
}
