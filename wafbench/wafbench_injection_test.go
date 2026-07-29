//go:build !libinjection_bench_default

package wafbench

import (
	"github.com/corazawaf/coraza/v3/experimental/plugins"
	"github.com/corazawaf/coraza/v3/experimental/plugins/plugintypes"

	"github.com/wasilibs/go-libinjection"
)

type detectSQLi struct{}

var _ plugintypes.Operator = (*detectSQLi)(nil)

func newDetectSQLi(plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	return &detectSQLi{}, nil
}

func (o *detectSQLi) Evaluate(tx plugintypes.TransactionState, value string) bool {
	res, fingerprint := libinjection.IsSQLi(value)
	if !res {
		return false
	}
	tx.CaptureField(0, string(fingerprint))
	return true
}

type detectXSS struct{}

var _ plugintypes.Operator = (*detectXSS)(nil)

func newDetectXSS(plugintypes.OperatorOptions) (plugintypes.Operator, error) {
	return &detectXSS{}, nil
}

func (o *detectXSS) Evaluate(_ plugintypes.TransactionState, value string) bool {
	return libinjection.IsXSS(value)
}

func init() {
	plugins.RegisterOperator("detectSQLi", newDetectSQLi)
	plugins.RegisterOperator("detectXSS", newDetectXSS)
}
