//go:build tinygo.wasm || libinjection_cgo

package cinjection

/*
int libinjection_sqli(void* s, int len, void* fp);
int libinjection_xss(void* s, int len);
// Prototypes for the libinjection functions linked from the wasm module (tinygo) or C library (cgo).
*/
import "C"

import "unsafe"

// As per https://github.com/libinjection/libinjection/blob/main/MIGRATION.md#strategy-2-minimal-migration-quick-fix
// Any returned value != LIBINJECTION_RESULT_FALSE (0) is either an injection or an error. Fail-safe approach: treat errors as detections.
func IsSQLi(sPtr unsafe.Pointer, sLen int, fpPtr unsafe.Pointer) bool {
	return C.libinjection_sqli(sPtr, C.int(sLen), fpPtr) != 0
}

func IsXSS(sPtr unsafe.Pointer, sLen int) bool {
	return C.libinjection_xss(sPtr, C.int(sLen)) != 0
}
