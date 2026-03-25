package main

import "kyla-2FA/internal"

func main() {
	internal.InitApp()
	internal.ExecuteUnlockFlow()
	internal.MainWindow.ShowAndRun()
}
