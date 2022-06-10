package skeleton

import "embed"

//go:embed handler/**/* lib/* static/* tests/**/* .gitignore kartusche.yaml init.js
var Content embed.FS
