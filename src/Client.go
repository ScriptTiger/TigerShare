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
		progress, chunkCount int
		displayProgress js.Value

		// WritableStreamDefaultWriter
		writer js.Value

		// IndexedDB
		db js.Value

		// Whether GetshowSaveFilePicker is available or not
		hasPicker bool
	)

	// Check if showSaveFilePicker is supported
	hasPicker = !jsGo.Get("showSaveFilePicker").IsUndefined()

	// Connect to signaling server
	peer = jsGo.Get("Peer").New(nil, getOptions())

	// Error event listener
	peer.Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
		errType := err[0].Get("type").String()
		if errType == "peer-unavailable" {unreachablePage()
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
		conn.Call("on", "close", jsGo.SimpleProcOf(func() {if !complete {unreachablePage()}}))

		// Start timeout timer
		jsGo.SetTimeout(jsGo.SimpleProcOf(func() {if !connected{unreachablePage()}}), 5000)

		// Data connection event listener
		conn.Call("on", "data", jsGo.ProcOf(func(data []js.Value) {

			// Update status
			connected = true

			// Function to display page while download in progress
			downloadingPage := func() {
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
			}

			// Function to update progress
			updateProgress := func() {
				progress += chunkSize
				if progress > fileSize {progress = fileSize}
				displayProgress.Set("textContent", jsGo.String.Invoke(progress).String()+" bytes out of "+jsGo.String.Invoke(fileSize).String())
			}

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

				// Download button for desktop with writeable stream
				if hasPicker {
					appAppendChild(
						jsGo.CreateSaveFileButton(
							"Download",

							// Set file name as suggested save name
							map[string]any{"suggestedName": fileName},

							// Receive saveFile from user and get WritableStreamDefaultWriter
							func(saveFile js.Value) {
								jsGo.ThenableChain(
									saveFile.Call("createWritable"),
									func(writeableStream js.Value) (any) {
										writer = writeableStream.Call("getWriter")
										downloadingPage()
										return nil
									},
								)
							},
						),
					)

				// Download button for mobile with indexedDB
				} else {
					appAppendChild(
						jsGo.CreateButton(
							"Download",
							func() {

								// Request to open a DB and capture resulting DB
								request := jsGo.IndexedDB.Call("open", pid+tid+fileName, 1)
								request.Set("onupgradeneeded", jsGo.ProcOf(func(event []js.Value) {
									event[0].Get("target").Get("result").Call(
										"createObjectStore",
										"chunks",
										map[string]any{"keyPath": "id",},
									)
								}))
								request.Set("onsuccess", jsGo.ProcOf(func(event []js.Value) {
									db = event[0].Get("target").Get("result")
									downloadingPage()
								}))
							},
						),
					)
				}

			// Receive file chunks and write them to disk
			} else {

				// Stream chunks directly to save file on desktop browsers
				if hasPicker {
					jsGo.ThenableChain(
						writer.Call("write", data[0]),

						// If chunk received and written successfully, update status
						func(succeed js.Value) (any) {
							updateProgress()

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
						func(fail js.Value) {tryAgainPage()},
					)

				// Store chunks in indexedDB on mobile browsers, and then reassemble chunks before writing to disk
				} else {
					chunkCount++
					transaction := db.Call("transaction", []any{"chunks"}, "readwrite")
					store := transaction.Call("objectStore", "chunks")
					request := store.Call("put", map[string]any{"id": chunkCount, "chunk": jsGo.Blob.New(jsGo.Array.New(data[0]))})
					request.Set("onerror", jsGo.SimpleProcOf(func() {tryAgainPage()}))
					updateProgress()

					// If last chunk received is last chunk of file, reassemble the file, download to disk, update page, and disconnect from everything
					if progress == fileSize {

						// Make sure last transaction is complete before proceeding
						transaction.Set("oncomplete", jsGo.SimpleProcOf(func() {

							// Create a new array to collect all of the chunks, and a new transaction to read the chunks into that array
							chunksArray := jsGo.Array.New()
							transaction := db.Call("transaction", []any{"chunks"}, "readonly")
							store := transaction.Call("objectStore", "chunks")

							// Begin iterating the records and reading the chunks into the array
							request := store.Call("openCursor", jsGo.IDBKeyRange.Call("bound", 1, chunkCount))
							request.Set("onsuccess", jsGo.ProcOf(func(event []js.Value) {
								cursor := event[0].Get("target").Get("result")

								// Read the chunk into the array
								if !cursor.IsUndefined() && !cursor.IsNull() {
									chunksArray.Call("push", cursor.Get("value").Get("chunk"))
									cursor.Call("continue")

								// If all records have been read, trigger the downloading of the combined blob, update the page, and clean up
								} else {
									jsGo.IndexedDB.Call("deleteDatabase", pid+tid+fileName)
									anchor := jsGo.CreateElement("a")
									anchor.Set("download", fileName)
									anchor.Set("href", jsGo.URL.Call("createObjectURL", jsGo.Blob.New(chunksArray)))
									anchor.Set("hidden", true)
									appAppendChild(anchor)
									anchor.Call("click")
									complete = true
									app.Set("innerHTML", nil)
									header := jsGo.CreateElement("h1")
									header.Set("textContent", "Your download has completed! However, your browser may still be saving the file.")
									appAppendChild(header)
									explainer := jsGo.CreateElement("h2")
									explainer.Set("textContent", "Please wait until your browser has finished saving the file before navigating away from this page.")
									appAppendChild(explainer)
									peer.Call("disconnect")
									peer.Call("destroy")
								}
							}))
						}))

					// If transfer not yet complete, send ACK for next chunk
					} else {conn.Call("send", "ACK")}
				}
			}
		}))

	}))
}
