//go:build !tinygo.wasm && !libinjection_cgo

package libinjection

import (
	"bytes"
	"sync"

	wasm2go "github.com/wasilibs/go-libinjection/internal/wasm"
)

// fingerprintSize is the size of the fingerprint buffer passed to
// libinjection_sqli, which is 8 characters plus a null terminator.
const fingerprintSize = 9

func IsSQLi(input string) (bool, string) {
	if len(input) == 0 {
		return false, ""
	}
	abi, _ := abiPool.Get().(*libinjectionABI)
	defer abiPool.Put(abi)

	abi.reserve(uint32(len(input) + fingerprintSize))
	sPtr := abi.writeString(input)
	fpPtr := abi.allocate(fingerprintSize)

	if abi.mod.Xlibinjection_sqli(int32(sPtr), int32(len(input)), int32(fpPtr)) == 0 {
		return false, ""
	}

	fpBuf := abi.read(fpPtr, fingerprintSize)
	nullIdx := bytes.IndexByte(fpBuf, 0)
	if nullIdx == -1 {
		nullIdx = fingerprintSize
	}
	fp := string(fpBuf[:nullIdx])

	return true, fp
}

func IsXSS(input string) bool {
	if len(input) == 0 {
		return false
	}
	abi, _ := abiPool.Get().(*libinjectionABI)
	defer abiPool.Put(abi)

	abi.reserve(uint32(len(input)))
	sPtr := abi.writeString(input)

	// Any returned value other than LIBINJECTION_RESULT_FALSE (0) is either an
	// injection or an error. To fail safe, errors are treated as detections, as
	// recommended by
	// https://github.com/libinjection/libinjection/blob/main/MIGRATION.md#strategy-2-minimal-migration-quick-fix
	return abi.mod.Xlibinjection_xss(int32(sPtr), int32(len(input))) != 0
}

var abiPool = sync.Pool{
	New: func() any {
		mod := wasm2go.New(wasm2go.NewHostWASI())
		mod.X_initialize()
		return &libinjectionABI{mod: mod}
	},
}

// libinjectionABI is a module instance along with a buffer in its memory that
// is reused across calls to pass input. Each instance has its own memory, so
// an instance must only be used by one goroutine at a time.
type libinjectionABI struct {
	mod *wasm2go.Module

	bufSize uint32
	bufPtr  uint32
	nextIdx uint32
}

// reserve ensures the buffer can hold size bytes and resets allocations.
func (abi *libinjectionABI) reserve(size uint32) {
	abi.nextIdx = 0
	if abi.bufSize >= size {
		return
	}

	if abi.bufPtr != 0 {
		abi.mod.Xfree(int32(abi.bufPtr))
		abi.bufPtr = 0
		abi.bufSize = 0
	}
	ptr := uint32(abi.mod.Xmalloc(int32(size)))
	if ptr == 0 {
		panic("failed to allocate wasm memory")
	}
	abi.bufPtr = ptr
	abi.bufSize = size
}

func (abi *libinjectionABI) allocate(size uint32) uint32 {
	if abi.nextIdx+size > abi.bufSize {
		panic("not enough reserved shared memory")
	}

	ptr := abi.bufPtr + abi.nextIdx
	abi.nextIdx += size
	return ptr
}

// memory returns the current module memory. It must be fetched again after any
// call into the module since memory growth may reallocate it.
func (abi *libinjectionABI) memory() []byte {
	return *abi.mod.Xmemory().Slice()
}

func (abi *libinjectionABI) read(ptr uint32, size uint32) []byte {
	return abi.memory()[ptr : ptr+size]
}

func (abi *libinjectionABI) writeString(s string) uint32 {
	ptr := abi.allocate(uint32(len(s)))
	copy(abi.memory()[ptr:], s)
	return ptr
}
