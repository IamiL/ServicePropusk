package main

import (
	appService "rip/app"
)

func main() {
	app := appService.New()

	app.MustRun()
}
