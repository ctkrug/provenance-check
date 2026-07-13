//go:build js && wasm

package main

import "syscall/js"

// main registers window.provenanceCheck(url) and then parks forever. Go's
// wasm scheduler yields back to the JS event loop once every goroutine is
// blocked, so this doesn't freeze the page — it just keeps the runtime
// (and its registered callbacks) alive.
func main() {
	js.Global().Set("provenanceCheck", js.FuncOf(checkOne))
	js.Global().Set("provenanceCheckReady", js.ValueOf(true))
	select {}
}

// checkOne is the JS-callable entry point: provenanceCheck(url) returns a
// Promise that resolves to the JSON-decoded checkResult for that url. The
// check itself runs on its own goroutine, so calling it once per pasted
// URL lets the browser resolve them concurrently and update each exhibit
// card as its own promise settles.
func checkOne(this js.Value, args []js.Value) any {
	promiseCtor := js.Global().Get("Promise")
	handler := js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve, reject := promiseArgs[0], promiseArgs[1]
		if len(args) < 1 {
			reject.Invoke(js.Global().Get("Error").New("provenanceCheck requires a url argument"))
			return nil
		}
		url := args[0].String()

		go func() {
			data, err := checkJSON(url)
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}
			resolve.Invoke(js.Global().Get("JSON").Call("parse", data))
		}()
		return nil
	})
	return promiseCtor.New(handler)
}
