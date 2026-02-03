//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// Append a child element to the app container
func appAppendChild(child js.Value) {app.Call("appendChild", child)}

// Encode a regular string as a base64 URL-safe string
func stringToUrl(str string) (string) {
	return jsGo.String.New(jsGo.String.New(jsGo.String.New(jsGo.Btoa(str)).Call("replaceAll", "=", "")).Call("replaceAll", "+", "-")).Call("replaceAll", "/", "_").String()
}

// Decode a base64 URL-safe string back to a regular string
func urlToString(str string) (string) {
	return jsGo.Atob(jsGo.String.New(jsGo.String.New(str).Call("replaceAll", "-", "+")).Call("replaceAll", "_", "/")).String()
}

// Return peer options, currently only used for passing ICE/TURN configuration if present
func getOptions() (map[string]any) {
	if turnUrl == "" || turnUser == "" || turnCred == "" {return map[string]any{}}
	if policy == "" {policy = "all"}
	return map[string]any{
		"config": map[string]any{
			"iceServers": []any{
				map[string]any{
					"urls": "turn:"+turnUrl,
					"username": turnUser,
					"credential": turnCred,
				},
			},
			"iceTransportPolicy": policy,
		},
	}
}

// Page displayed when download link is unreachable
func unreachablePage() {
	app.Set("innerHTML", nil)
	header := jsGo.CreateElement("h1")
	header.Set("textContent", "The link is unreachable!")
	appAppendChild(header)
	peer.Call("disconnect")
	peer.Call("destroy")
}

// Page displayed when download interrupted 
func tryAgainPage() {
	app.Set("innerHTML", nil)
	header := jsGo.CreateElement("h1")
	header.Set("textContent", "Your download has failed!")
	appAppendChild(header)
	appAppendChild(jsGo.CreateButton("Try again", func() {jsGo.Location.Set("href", urlOrigin)}))
	peer.Call("disconnect")
	peer.Call("destroy")
}
