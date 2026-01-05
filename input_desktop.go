//go:build !js && !wasm

package main

func isMobileBrowser() bool {
	return false // On Mac, it's never a mobile browser
}
