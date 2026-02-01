//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// Connect to signaling server and download file
func client() {

	var (
		// Status tracking
		complete bool
		progress int
		displayProgress js.Value

		// WritableStreamDefaultWriter
		writer js.Value

		// Whether GetshowSaveFilePicker is available or not
		hasPicker bool
	)

	// Check if showSaveFilePicker is supported
	hasPicker = !jsGo.Get("showSaveFilePicker").IsUndefined()

	// If showSaveFilePicker is not supported, display notice and disconnect from everything
	if !hasPicker {
		app.Set("innerHTML", nil)
		header := jsGo.CreateElement("h1")
		header.Set("textContent", "Mobile browsers are not currently supported, but they will be soon!")
		appAppendChild(header)
		peer.Call("disconnect")
		peer.Call("destroy")
		return
	}

	// Connect to signaling server
	peer = jsGo.Get("Peer").New(nil, getOptions())

	// Error event listener
	peer.Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
		errType := err[0].Get("type").String()
		if errType == "peer-unavailable" {unreachable()
		} else {jsGo.Log(errType)}
	}))

	// Open event listener
	peer.Call("on", "open", jsGo.ProcOf(func(id []js.Value) {

		// Connect to server
		conn := peer.Call("connect", pid, map[string]any{
			"label": tid,
			"serialization": "raw",
			"reliable": true,
		})

		// Error connection event listener
		conn.Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
			jsGo.Log("Data connection error with "+conn.Get("peer").String()+": "+err[0].Get("type").String())
		}))

		// Close connection event listener
		conn.Call("on", "close", jsGo.SimpleProcOf(func() {if !complete {unreachable()}}))

		// Start timeout timer
		jsGo.SetTimeout(jsGo.SimpleProcOf(func() {if !connected{unreachable()}}), 5000)

		// Data connection event listener
		conn.Call("on", "data", jsGo.ProcOf(func(data []js.Value) {

			// Update status
			connected = true

			// Get file name from server
			if fileName == "" {
				fileName = data[0].String()
				conn.Call("send", "ACK")
				return
			}

			// Get file size from server, and then present file info and download button
			if fileSize == 0 {
				fileSize = jsGo.Number.Invoke(data[0]).Int()

				// Wipe app area
				app.Set("innerHTML", nil)

				// File info
				fileInfo := jsGo.CreateElement("p")
				fileInfo.Set("textContent", "\""+fileName+"\" ("+jsGo.String.Invoke(fileSize).String()+" bytes)")
				appAppendChild(fileInfo)

				// Download button
				downloadButton := jsGo.CreateSaveFileButton(
					"Download",

					// Set file name as suggested save name
					map[string]any{"suggestedName": fileName},

					// Receive saveFile from user and get WritableStreamDefaultWriter
					func(saveFile js.Value) {
						jsGo.ThenableChain(
							saveFile.Call("createWritable"),
							func(writeableStream js.Value) (any) {
								writer = writeableStream.Call("getWriter")

								// Set up explainer
								app.Set("innerHTML", nil)
								header := jsGo.CreateElement("h1")
								header.Set("textContent", "You are currently downloading in the background.")
								appAppendChild(header)
								explainer := jsGo.CreateElement("h2")
								explainer.Set("textContent", "Do not navigate away from this page until the download has completed.")
								appAppendChild(explainer)
								appAppendChild(jsGo.CreateElement("br"))
								appAppendChild(jsGo.CreateElement("br"))
								appAppendChild(jsGo.CreateElement("br"))

								// Set up element to display progress
								displayProgress = jsGo.CreateElement("h2")
								displayProgress.Set("textContent", "0 bytes out of "+jsGo.String.Invoke(fileSize).String())
								appAppendChild(displayProgress)

								// Signal server to start sending file
								conn.Call("send", "ACK")

								return nil
							},
						)
					},
				)
				appAppendChild(downloadButton)

			// Receive file chunks and stream them into the selected saveFile
			} else {
				jsGo.ThenableChain(
					writer.Call("write", data[0]),

					// If chunk received and written successfully, update status accordingly
					func(succeed js.Value) (any) {
						progress += chunkSize
						if progress > fileSize {progress = fileSize}
						displayProgress.Set("textContent", jsGo.String.Invoke(progress).String()+" bytes out of "+jsGo.String.Invoke(fileSize).String())

						// If last chunk received is last chunk of file, update page and disconnect from everything
						if progress == fileSize {
							writer.Call("close")
							complete = true
							app.Set("innerHTML", nil)
							header := jsGo.CreateElement("h1")
							header.Set("textContent", "Your download has completed!")
							appAppendChild(header)
							explainer := jsGo.CreateElement("h2")
							explainer.Set("textContent", "You may now safely navigate away from this page.")
							appAppendChild(explainer)
							peer.Call("disconnect")
							peer.Call("destroy")

						// If transfer not yet complete, send ACK for next chunk
						} else {conn.Call("send", "ACK")}
						return nil
					},

					// If there was a problem receiving and writing chunk, update page and disconnect from everything
					func(fail js.Value) {
						app.Set("innerHTML", nil)
						header := jsGo.CreateElement("h1")
						header.Set("textContent", "Your download has failed!")
						appAppendChild(header)
						appAppendChild(jsGo.CreateButton("Try again", func() {jsGo.Location.Set("href", urlOrigin)}))
						peer.Call("disconnect")
						peer.Call("destroy")
					},
				)
			}
		}))

	}))
}
