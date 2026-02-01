//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// Connect to signaling server and serve file to connecting clients
func server() {

	var (
		// Status tracking
		connections, shares int
		displayConnections, displayShares js.Value
	)

	// Connect to signalling server
	peer = jsGo.Get("Peer").New(nil, getOptions())

	// Update status
	connected = false
	destroyed = false

	// Error event listener
	peer.Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
		jsGo.Log(err[0].Get("type").String())
	}))

	// Disconnected event listener
	peer.Call("on", "disconnected", jsGo.SimpleProcOf(func() {

		// Update state
		connected = false

		if !destroyed {peer.Call("reconnect")
		} else {connected = false}
	}))

	// Close event listener
	peer.Call("on", "close", jsGo.SimpleProcOf(func() {
		connected = false
		destroyed = true
	}))

	// Open event listener
	peer.Call("on", "open", jsGo.ProcOf(func(id []js.Value) {

		// Update status
		connected = true
		destroyed = false

		// Set up page
		if !hasPage {

			// Update status
			hasPage = true

			// Wipe current app area
			app.Set("innerHTML", nil)

			// Explainer text while sharing
			header := jsGo.CreateElement("h1")
			header.Set("textContent", "You are currently sharing.")
			appAppendChild(header)
			explainer := jsGo.CreateElement("h2")
			explainer.Set("innerHTML",
				"If you navigate away from this page or click the \"Stop sharing\" button, you will stop sharing and your link will become invalid.<br>"+
				"A new link is generated each time you share a file.<br><br><br>",
			)
			appAppendChild(explainer)

			// Stats
			displayConnections = jsGo.CreateElement("h3")
			displayConnections.Set("textContent", "Active connections: 0")
			appAppendChild(displayConnections)
			displayShares = jsGo.CreateElement("h3")
			displayShares.Set("textContent", "Completed shares: 0")
			appAppendChild(displayShares)
			appAppendChild(jsGo.CreateElement("br"))
			appAppendChild(jsGo.CreateElement("br"))
			appAppendChild(jsGo.CreateElement("br"))

			// Capture PID
			pid = id[0].String()

			// Generate TID
			jsTID := jsGo.Uint8Array.New(16)
			jsGo.Crypto.Call("getRandomValues", jsTID)
			tid = jsGo.String.New(jsGo.String.New(jsGo.String.New(jsTID.Call("toBase64")).Call("replaceAll", "=", "")).Call("replaceAll", "+", "-")).Call("replaceAll", "/", "_").String()

			// Stop button to disconnect from everything and destroy peer object
			appAppendChild(
				jsGo.CreateButton("Stop sharing", func() {
					connected = false
					destroyed = true
					peer.Call("disconnect")
					peer.Call("destroy")
					app.Set("innerHTML", nil)
					appAppendChild(jsGo.CreateButton("Share another file", func() {jsGo.Location.Set("href", urlFull)}))
				}),
			)

			// Share link for share button and QR code
			shareLink := url+"?pid="+pid+"&tid="+tid

			// Share button to copy share link to clipboard
			appAppendChild(
				jsGo.CreateButton("Copy share link", func() {
					jsGo.Get("navigator").Get("clipboard").Call("writeText", shareLink)
				}),
			)

			// Display QR code with share link
			qrCode := jsGo.CreateElement("div")
			appAppendChild(qrCode)
			jsGo.Get("QRCode").New(qrCode, shareLink)
		}
	}))

	// Connection event listener
	peer.Call("on", "connection", jsGo.ProcOf(func(conn []js.Value) {

		// Status tracking
		var progress, end int

		if conn[0].Get("label").String() == tid {

			// Error connection event listener
			conn[0].Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
				jsGo.Log("Data connection error with "+conn[0].Get("peer").String()+": "+err[0].Get("type").String())
			}))

			// Close connection event listener
			conn[0].Call("on", "close", jsGo.SimpleProcOf(func() {

				// Update stats
				connections--
				displayConnections.Set("textContent", "Active connections: "+jsGo.String.Invoke(connections).String())
			}))

			// Open connection event listener
			conn[0].Call("on", "open", jsGo.SimpleProcOf(func() {

				// Update stats
				connections++
				displayConnections.Set("textContent", "Active connections: "+jsGo.String.Invoke(connections).String())

				// Send file name first
				conn[0].Call("send", fileName)
			}))

			// Data connection event listener
			conn[0].Call("on", "data", jsGo.ProcOf(func(data []js.Value) {

				// If this is the first ACK received, calculate first chunk and send file size
				if progress == 0 && end == 0 {

					// Calculate end of first chunk
					end = chunkSize
					if end > fileSize  {end = fileSize}

					// Send file size
					conn[0].Call("send", fileSize)
					return
				}

				// Reply to subsequent ACKs with file chunks
				if data[0].String() == "ACK" {

					// Only send a chunk if transfer not yet complete, otherwise don't send anything
					if end != 0 {

						// Get chunk slice
						slice := file.Call("slice", progress, end)

						// Get chunk array buffer and send it
						jsGo.ThenableChain(
							slice.Call("arrayBuffer"),
							func(arrayBuffer js.Value) (any) {
								conn[0].Call("send", arrayBuffer)

								// If the entire file has been sent, zero end, start timer to disconnect, and update stats
								if end == fileSize {
									end = 0
									jsGo.SetTimeout(jsGo.SimpleProcOf(func() {conn[0].Call("close")}), 5000)
									shares++
									displayShares.Set("textContent", "Completed shares: "+jsGo.String.Invoke(shares).String())

								// If more chunks need to be sent, calculate the next chunk
								} else {
									progress += chunkSize
									end += chunkSize
									if end > fileSize {end = fileSize}
								}
								return nil
							},
						)
					}
				}
			}))

		// Close connection if incorrect TID
		} else {conn[0].Call("close")}
	}))
}
