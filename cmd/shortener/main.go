package main

import "github.com/hollgett/shortener.git/internal/app"

func main() {
	app := app.NewApp()
	app.Run()
}
