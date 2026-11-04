//nolint:revive // Method names must match the wasm2go generated interface.
package wasm2go

import (
	"encoding/binary"
	"io"
	"os"
)

const (
	errnoSuccess = 0
	errnoBadf    = 8
	errnoIo      = 29
	errnoSpipe   = 70
)

// HostWASI is a minimal implementation of the wasi_snapshot_preview1 imports
// used by libinjection. libinjection does not perform I/O itself, so only
// writes to stdout and stderr, as may happen on abort, are supported.
//
// A HostWASI must not be shared between modules since it is bound to the
// memory of the module it is passed to.
type HostWASI struct {
	stdout io.Writer
	stderr io.Writer

	module *Module
}

// NewHostWASI returns a HostWASI writing to the process stdout and stderr.
func NewHostWASI() *HostWASI {
	return &HostWASI{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Init binds the host to the memory of module. Callers do not need to call it
// because New calls it.
func (w *HostWASI) Init(module any) {
	if m, ok := module.(*Module); ok {
		w.module = m
	}
}

func (w *HostWASI) Xfd_close(v0 int32) int32 {
	switch v0 {
	case 0, 1, 2:
		return errnoSuccess
	default:
		return errnoBadf
	}
}

func (w *HostWASI) Xfd_seek(int32, int64, int32, int32) int32 {
	return errnoSpipe
}

func (w *HostWASI) Xfd_write(v0, v1, v2, v3 int32) int32 {
	var out io.Writer
	switch v0 {
	case 1:
		out = w.stdout
	case 2:
		out = w.stderr
	default:
		return errnoBadf
	}

	mem := w.module.memory
	var total uint32
	errno := int32(errnoSuccess)
	for i := range v2 {
		iovec := uint32(v1 + i*8)
		bufPtr := binary.LittleEndian.Uint32(mem[iovec:])
		bufLen := binary.LittleEndian.Uint32(mem[iovec+4:])
		chunk := mem[bufPtr : bufPtr+bufLen]

		n, err := out.Write(chunk)
		total += uint32(n)
		if err != nil || n != len(chunk) {
			errno = errnoIo
			break
		}
	}

	binary.LittleEndian.PutUint32(mem[uint32(v3):], total)
	return errno
}

var _ Xwasi_snapshot_preview1 = (*HostWASI)(nil)
