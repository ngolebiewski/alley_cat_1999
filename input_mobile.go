//go:build js || wasm

package main

import (
	"strings"
	"syscall/js"
)

func isMobileBrowser() bool {
	ua := js.Global().Get("navigator").Get("userAgent").String()
	ua = strings.ToLower(ua)

	// Standard checks for Android and iPhone
	if strings.Contains(ua, "android") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		return true
	}

	// Modern iPad Check:
	// It reports as "Macintosh" but has multi-touch support.
	maxTouchPoints := js.Global().Get("navigator").Get("maxTouchPoints").Int()
	isMac := strings.Contains(ua, "macintosh")

	return isMac && maxTouchPoints > 1
}

// package main

// import (
// 	"runtime"
// 	"strings"
// 	"syscall/js"
// )

// // This checks if we are on a mobile device and then should add in a virtual joystick

// var isMobile = false

// func init() {
// 	// Check if we are in a browser (WASM)
// 	if runtime.GOOS == "js" {
// 		ua := js.Global().Get("navigator").Get("userAgent").String()
// 		ua = strings.ToLower(ua)
// 		isMobile = strings.Contains(ua, "android") ||
// 			strings.Contains(ua, "iphone") ||
// 			strings.Contains(ua, "ipad")
// 	} else if runtime.GOOS == "android" || runtime.GOOS == "ios" {
// 		// If you ever build natively for mobile
// 		isMobile = true
// 	}
// }
