package skeleton

import "embed"

//go:embed handler/**/* lib/* static/* tests/**/* .gitignore init.js cronjobs/*
var Content embed.FS
