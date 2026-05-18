package main

import "embed"

//go:embed ui/dist
var StaticFS embed.FS
