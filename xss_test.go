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
	// TAB, LF and CR inside a URL scheme, which browsers strip (GHSA-r97g-xfh2-h3jj).
	"<a href=\"ja\tvascript:alert(1)\">x</a>",
	"<a href=\"ja\rvascript:alert(1)\">x</a>",
	"<a href=\"ja&#9;vascript:alert(1)\">x</a>",
	"<a href=\"ja&#x09;vascript:alert(1)\">x</a>",
	"<a href=\"ja&#10;vascript:alert(1)\">x</a>",
	"<a href=\"ja&#13;vascript:alert(1)\">x</a>",
	"<a href=\"ja&Tab;vascript:alert(1)\">x</a>",
	"<a href=\"ja&NewLine;vascript:alert(1)\">x</a>",
	"<a href=\"&Tab;javascript:alert(1)\">x</a>",
	"<form action=\"ja&Tab;vascript:alert(1)\">",
	"<button formaction=\"ja&#9;vascript:alert(1)\">x</button>",
	"<a href=\"vb&#9;script:x\">x</a>",
	"<a href=\"da&#9;ta:text/html,x\">x</a>",
	"<a href=\"vb&NewLine;script:x\">x</a>",
	"<a href=\"view-so&Tab;urce:http://x\">x</a>",
	"<a href=\"ja&Tab;&Tab;vascript:alert(1)\">x</a>",
}

var notXSSExamples = []string{
	// A reference above 0xFF is no letter: &#330; must not truncate to 'J'.
	"<a href=\"&#330;avascript:alert(1)\">x</a>",
	// Browsers strip TAB, LF and CR in any URL, not only dangerous ones.
	"<a href=\"http://ex&#9;ample.com/&Tab;\">x</a>",
	"<a href=\"ht&Tab;tp://x\">x</a>",
	// Only &Tab; and &NewLine; are decoded, case-sensitively and with the ';'.
	"<a href=\"ja&tab;vascript:alert(1)\">x</a>",
	"<a href=\"ja&Tabvascript:alert(1)\">x</a>",
	// A vertical tab is not stripped.
	"<a href=\"ja\vvascript:alert(1)\">x</a>",
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

func TestIsNotXSS(t *testing.T) {
	for _, example := range notXSSExamples {
		t.Run(example, func(t *testing.T) {
			if IsXSS(example) {
				t.Errorf("[%s] is XSS", example)
			}
		})
	}
}

func BenchmarkIsXSS(b *testing.B) {
	for _, example := range xssExamples {
		b.Run(example, func(b *testing.B) {
			for range b.N {
				if !IsXSSBenchmark(example) {
					b.Errorf("[%s] is not XSS", example)
				}
			}
		})
	}
}
