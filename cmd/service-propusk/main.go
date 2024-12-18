package main

import (
	appService "rip/internal/app"
)

func main() {
	app := appService.New()

	app.MustRun()
}
