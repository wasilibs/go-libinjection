//go:build !tinygo.wasm && !libinjection_cgo

package libinjection

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var (
	errFailedRead = errors.New("failed to read from wasm memory")
)

//go:embed wasm/libinjection.so
var libinjection []byte

var (
	wasmRT       wazero.Runtime
	wasmCompiled wazero.CompiledModule
)

func init() {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	code, err := rt.CompileModule(ctx, libinjection)
	if err != nil {
		panic(err)
	}
	wasmCompiled = code

	wasmRT = rt
}

func IsSQLi(input string) (bool, string) {
	if len(input) == 0 {
		return false, ""
	}
	abi := abiPool.Get().(*libinjectionABI)
	defer abiPool.Put(abi)

	abi.startOperation(len(input) + 9)
	defer abi.endOperation()
	sPtr := abi.memory.writeString(abi, input)
	fpPtr := abi.memory.allocate(9)

	res, err := abi.libinjectionSQLi.Call(context.Background(), uint64(sPtr), uint64(len(input)), uint64(fpPtr))
	if err != nil {
		panic(err)
	}

	fpBuf := abi.memory.read(abi, fpPtr, 9)
	nullIdx := bytes.IndexByte(fpBuf, 0)
	if nullIdx == -1 {
		nullIdx = 9
	}
	fp := string(fpBuf[:nullIdx])

	return res[0] == 1, fp
}

func IsXSS(input string) bool {
	if len(input) == 0 {
		return false
	}
	abi := abiPool.Get().(*libinjectionABI)
	defer abiPool.Put(abi)

	abi.startOperation(len(input))
	defer abi.endOperation()
	sPtr := abi.memory.writeString(abi, input)

	res, err := abi.libinjectionXSS.Call(context.Background(), uint64(sPtr), uint64(len(input)))
	if err != nil {
		panic(err)
	}

	return res[0] == 1
}

var (
	moduleIdx = uint64(0)
	abiPool   = sync.Pool{
		New: func() interface{} {
			ctx := context.Background()
			modIdx := atomic.AddUint64(&moduleIdx, 1)
			mod, err := wasmRT.InstantiateModule(ctx, wasmCompiled, wazero.NewModuleConfig().WithName(strconv.FormatUint(modIdx, 10)))
			if err != nil {
				panic(err)
			}

			abi := &libinjectionABI{
				libinjectionSQLi: mod.ExportedFunction("libinjection_sqli"),
				libinjectionXSS:  mod.ExportedFunction("libinjection_xss"),
				malloc:           mod.ExportedFunction("malloc"),
				free:             mod.ExportedFunction("free"),

				wasmMemory: mod.Memory(),
				mod:        mod,
			}

			return abi
		},
	}
)

type libinjectionABI struct {
	libinjectionSQLi api.Function
	libinjectionXSS  api.Function

	malloc api.Function
	free   api.Function

	wasmMemory api.Memory

	mod api.Module

	memory sharedMemory
}

func (abi *libinjectionABI) startOperation(memorySize int) {
	abi.memory.reserve(abi, uint32(memorySize))
}

func (abi *libinjectionABI) endOperation() {
}

type sharedMemory struct {
	size    uint32
	bufPtr  uint32
	nextIdx uint32
}

func (m *sharedMemory) reserve(abi *libinjectionABI, size uint32) {
	m.nextIdx = 0
	if m.size >= size {
		return
	}

	ctx := context.Background()
	if m.bufPtr != 0 {
		_, err := abi.free.Call(ctx, uint64(m.bufPtr))
		if err != nil {
			panic(err)
		}
	}

	res, err := abi.malloc.Call(ctx, uint64(size))
	if err != nil {
		panic(err)
	}

	m.size = size
	m.bufPtr = uint32(res[0])
}

func (m *sharedMemory) allocate(size uint32) uintptr {
	if m.nextIdx+size > m.size {
		panic("not enough reserved shared memory")
	}

	ptr := m.bufPtr + m.nextIdx
	m.nextIdx += size
	return uintptr(ptr)
}

func (m *sharedMemory) read(abi *libinjectionABI, ptr uintptr, size int) []byte {
	buf, ok := abi.wasmMemory.Read(uint32(ptr), uint32(size))
	if !ok {
		panic(errFailedRead)
	}
	return buf
}

func (m *sharedMemory) writeString(abi *libinjectionABI, s string) uintptr {
	ptr := m.allocate(uint32(len(s)))
	abi.wasmMemory.WriteString(uint32(ptr), s)
	return ptr
}
