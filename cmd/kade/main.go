package main

import (
	"github.com/adde/kade/internal/app"
)

func main() {
	app.ParseFlags()
	app.Create()
}
