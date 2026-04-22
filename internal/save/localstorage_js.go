//go:build js && wasm

package save

import "syscall/js"

var storage = js.Global().Get("localStorage")

func Write(key, value string) { storage.Call("setItem", key, value) }

func Read(key string) (string, bool) {
	v := storage.Call("getItem", key)
	if v.IsNull() || v.IsUndefined() {
		return "", false
	}
	return v.String(), true
}
