//go:build tinygo.wasm

package libinjection

/*
#cgo LDFLAGS: -Linternal/wasm -linjection
*/
import "C"
