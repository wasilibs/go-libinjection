package libinjection

import (
	"testing"
)

var xssExamples = []string{
	"<script>alert(1);</script>",
	"><script>alert(1);</script>",
	"x ><script>alert(1);</script>",
	"' ><script>alert(1);</script>",
	"\"><script>alert(1);</script>",
	"red;</style><script>alert(1);</script>",
	"red;}</style><script>alert(1);</script>",
	"red;\"/><script>alert(1);</script>",
	"');}</style><script>alert(1);</script>",
	"onerror=alert(1)>",
	"x onerror=alert(1);>",
	"x' onerror=alert(1);>",
	"x\" onerror=alert(1);>",
	"<a href=\"javascript:alert(1)\">",
	"<a href='javascript:alert(1)'>",
	"<a href=javascript:alert(1)>",
	"<a href  =   javascript:alert(1); >",
	"<a href=\"  javascript:alert(1);\" >",
	"<a href=\"JAVASCRIPT:alert(1);\" >",
}

func TestIsXSS(t *testing.T) {
	for _, example := range xssExamples {
		t.Run(example, func(t *testing.T) {
			if !IsXSS(example) {
				t.Errorf("[%s] is not XSS", example)
			}
		})
	}
}

func BenchmarkIsXSS(b *testing.B) {
	for _, example := range xssExamples {
		b.Run(example, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if !IsXSSBenchmark(example) {
					b.Errorf("[%s] is not XSS", example)
				}
			}
		})
	}
}
