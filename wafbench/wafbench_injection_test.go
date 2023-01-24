//go:build !libinjection_bench_default

package wafbench

import (
	"github.com/corazawaf/coraza/v3/operators"
	"github.com/corazawaf/coraza/v3/rules"

	"github.com/wasilibs/go-libinjection"
)

type detectSQLi struct{}

var _ rules.Operator = (*detectSQLi)(nil)

func newDetectSQLi(rules.OperatorOptions) (rules.Operator, error) {
	return &detectSQLi{}, nil
}

func (o *detectSQLi) Evaluate(tx rules.TransactionState, value string) bool {
	res, fingerprint := libinjection.IsSQLi(value)
	if !res {
		return false
	}
	tx.CaptureField(0, string(fingerprint))
	return true
}

type detectXSS struct{}

var _ rules.Operator = (*detectXSS)(nil)

func newDetectXSS(rules.OperatorOptions) (rules.Operator, error) {
	return &detectXSS{}, nil
}

func (o *detectXSS) Evaluate(_ rules.TransactionState, value string) bool {
	return libinjection.IsXSS(value)
}

func init() {
	operators.Register("detectSQLi", newDetectSQLi)
	operators.Register("detectXSS", newDetectXSS)
}
