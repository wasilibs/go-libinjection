//go:build libinjection_bench_default

package libinjection

import (
	base "github.com/corazawaf/libinjection-go"
)

func IsSQLiBenchmark(s string) (bool, string) {
	return base.IsSQLi(s)
}

func IsXSSBenchmark(s string) bool {
	return base.IsXSS(s)
}
