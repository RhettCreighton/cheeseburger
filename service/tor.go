package service

import (
	_ "embed"
)

//go:embed tor-bins/tor-linux-x86_64
var torBinary []byte
