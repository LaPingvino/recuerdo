//go:build js && wasm

// Command recuerdo-web is Recuerdo's web version: Go compiled to
// WebAssembly (GOOS=js GOARCH=wasm; scripts/build-web.sh). It offers
// internal/webapi to the page as globalThis.recuerdo; each function
// returns a JSON string ({"error": ...} on failure), except save, which
// returns the file's bytes as a Uint8Array.
package main

import (
	"encoding/json"
	"io"
	"log"
	"syscall/js"

	"github.com/LaPingvino/recuerdo/internal/webapi"
)

var app webapi.App

func reply(v any, err error) any {
	if err != nil {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		return string(b)
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func main() {
	log.SetOutput(io.Discard) // the lesson package's progress messages
	api := map[string]any{
		"open": js.FuncOf(func(_ js.Value, args []js.Value) any {
			data := make([]byte, args[1].Get("length").Int())
			js.CopyBytesToGo(data, args[1])
			return reply(app.Open(args[0].String(), data))
		}),
		"openText": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(app.OpenText(args[0].String(), args[1].String()))
		}),
		"lesson":  js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Lesson()) }),
		"choices": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Choices(), nil) }),
		"start": js.FuncOf(func(_ js.Value, args []js.Value) any {
			var o webapi.Options
			if len(args) > 0 {
				if err := json.Unmarshal([]byte(args[0].String()), &o); err != nil {
					return reply(nil, err)
				}
			}
			return reply(app.Start(o))
		}),
		"state": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.State(), nil) }),
		"answer": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(app.Answer(args[0].String()))
		}),
		"stop":   js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Stop(), nil) }),
		"report": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Report(), nil) }),
		"save": js.FuncOf(func(_ js.Value, args []js.Value) any {
			b, err := app.Save(args[0].String())
			if err != nil {
				return reply(nil, err)
			}
			out := js.Global().Get("Uint8Array").New(len(b))
			js.CopyBytesToJS(out, b)
			return out
		}),
	}
	js.Global().Set("recuerdo", js.ValueOf(api))
	// recuerdo is ready as soon as go.run returns; pages may also listen
	if d := js.Global().Get("dispatchEvent"); d.Type() == js.TypeFunction {
		js.Global().Call("dispatchEvent", js.Global().Get("Event").New("recuerdo-ready"))
	}
	select {} // keep the functions alive
}
