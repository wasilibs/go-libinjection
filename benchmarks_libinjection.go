//go:build !libinjection_bench_default

package libinjection

func IsSQLiBenchmark(s string) (bool, string) {
	return IsSQLi(s)
}

func IsXSSBenchmark(s string) bool {
	return IsXSS(s)
}
