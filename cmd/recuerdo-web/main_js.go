//go:build js && wasm

// Command recuerdo-web is Recuerdo's web version: Go compiled to
// WebAssembly (GOOS=js GOARCH=wasm; scripts/build-web.sh). It offers
// internal/webapi to the page as globalThis.recuerdo; each function
// returns a JSON string ({"error": ...} on failure), except save, which
// returns the file's bytes as a Uint8Array.
package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"syscall/js"

	"github.com/LaPingvino/recuerdo/internal/i18n"
	"github.com/LaPingvino/recuerdo/internal/langcode"
	"github.com/LaPingvino/recuerdo/internal/richtext"
	"github.com/LaPingvino/recuerdo/internal/typing"
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

// promise runs f in a goroutine and returns a Promise of its reply: work
// that waits on the network (fetch) cannot run in the page's event.
func promise(f func(ctx context.Context) (any, error)) js.Value {
	var handler js.Func
	handler = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resolve := args[0]
		go func() {
			defer handler.Release()
			resolve.Invoke(reply(f(context.Background())))
		}()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
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
		"lesson":   js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Lesson()) }),
		"wordList": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.WordList()) }),
		"choices":  js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Choices(), nil) }),
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
		"stop": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Stop(), nil) }),
		"setTitle": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(nil, app.SetTitle(args[0].String()))
		}),
		"addItem": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(app.AddItem(args[0].String(), args[1].String()))
		}),
		"updateItem": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(nil, app.UpdateItem(args[0].Int(), args[1].String(), args[2].String()))
		}),
		"removeItem": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(nil, app.RemoveItem(args[0].Int()))
		}),
		// map lessons (webapi/topo.go)
		"newTopo": js.FuncOf(func(_ js.Value, args []js.Value) any {
			data := make([]byte, args[1].Get("length").Int())
			js.CopyBytesToGo(data, args[1])
			return reply(app.NewTopo(args[0].String(), data))
		}),
		"addPlace": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(app.AddPlace(args[0].String(), args[1].Int(), args[2].Int()))
		}),
		"setPlace": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(nil, app.SetPlace(args[0].Int(), args[1].String(), args[2].Int(), args[3].Int()))
		}),
		"newMedia": js.FuncOf(func(_ js.Value, args []js.Value) any { return reply(app.NewMedia(args[0].String())) }),
		"addMediaFile": js.FuncOf(func(_ js.Value, args []js.Value) any {
			data := make([]byte, args[1].Get("length").Int())
			js.CopyBytesToGo(data, args[1])
			return reply(app.AddMediaFile(args[0].String(), data))
		}),
		"addMediaAddress": js.FuncOf(func(_ js.Value, args []js.Value) any { return reply(app.AddMediaAddress(args[0].String())) }),
		"setMedia": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(nil, app.SetMedia(args[0].Int(), args[1].String(), args[2].String(), args[3].String()))
		}),
		// mediaFile is an embedded file as {bytes: Uint8Array, type}, or null
		"mediaFile": js.FuncOf(func(_ js.Value, args []js.Value) any {
			b, typ := app.MediaFile(args[0].String())
			if b == nil {
				return js.Null()
			}
			arr := js.Global().Get("Uint8Array").New(len(b))
			js.CopyBytesToJS(arr, b)
			return map[string]any{"bytes": arr, "type": typ}
		}),
		"osmStyles": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.OSMStyles(), nil) }),
		"osmSearch": js.FuncOf(func(_ js.Value, args []js.Value) any {
			q := args[0].String()
			return promise(func(ctx context.Context) (any, error) { return app.OSMSearch(ctx, q) })
		}),
		"osmMap": js.FuncOf(func(_ js.Value, args []js.Value) any {
			i, style, title := args[0].Int(), args[1].Int(), args[2].String()
			return promise(func(ctx context.Context) (any, error) { return app.OSMMap(ctx, i, style, title) })
		}),
		"osmPlace": js.FuncOf(func(_ js.Value, args []js.Value) any {
			name := args[0].String()
			return promise(func(ctx context.Context) (any, error) { return app.OSMPlace(ctx, name) })
		}),
		"nearestPlace": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return app.NearestPlace(args[0].Int(), args[1].Int(), args[2].Float())
		}),
		"viewAnswer":  js.FuncOf(func(js.Value, []js.Value) any { return reply(app.ViewAnswer()) }),
		"judge":       js.FuncOf(func(_ js.Value, args []js.Value) any { return reply(nil, app.Judge(args[0].Bool())) }),
		"skip":        js.FuncOf(func(js.Value, []js.Value) any { return reply(nil, app.Skip()) }),
		"correctLast": js.FuncOf(func(js.Value, []js.Value) any { return reply(nil, app.CorrectLast()) }),
		// translations: the page puts <lang>.po files in the memory file
		// system first; t returns a plain string
		"setLanguage": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(nil, webapi.SetLanguage(args[0].String(), args[1].String()))
		}),
		"t":            js.FuncOf(func(_ js.Value, args []js.Value) any { return i18n.T(args[0].String()) }),
		"languageName": js.FuncOf(func(_ js.Value, args []js.Value) any { return i18n.Name(args[0].String()) }),
		"rightToLeft":  js.FuncOf(func(_ js.Value, args []js.Value) any { return i18n.RightToLeft(args[0].String()) }),
		// languageCode is the code of a language name ("Dutch": "nl"), for
		// the browser's spell checker
		"languageCode": js.FuncOf(func(_ js.Value, args []js.Value) any { return langcode.Guess(args[0].String()) }),
		// sanitize makes words with markup safe to show (internal/richtext)
		"sanitize": js.FuncOf(func(_ js.Value, args []js.Value) any { return richtext.Sanitize(args[0].String()) }),
		"report":   js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Report(), nil) }),
		// the typing course: profiles go in and out as JSON
		"typingChoices": js.FuncOf(func(js.Value, []js.Value) any { return reply(webapi.TypingChoices(), nil) }),
		"typingNew": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(webapi.TypingNew(args[0].String(), args[1].String(), args[2].String()))
		}),
		"typingShow": js.FuncOf(func(_ js.Value, args []js.Value) any {
			var p typing.Profile
			if err := json.Unmarshal([]byte(args[0].String()), &p); err != nil {
				return reply(nil, err)
			}
			return reply(webapi.TypingShow(p), nil)
		}),
		"typingFinish": js.FuncOf(func(_ js.Value, args []js.Value) any {
			var p typing.Profile
			if err := json.Unmarshal([]byte(args[0].String()), &p); err != nil {
				return reply(nil, err)
			}
			return reply(webapi.TypingFinish(p, args[1].Float(), args[2].Int()), nil)
		}),
		"palette": js.FuncOf(func(js.Value, []js.Value) any { return reply(app.Palette(), nil) }),
		"expand": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return reply(app.Expand(args[0].String(), args[1].String()))
		}),
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
