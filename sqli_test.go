package libinjection

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestIsSQLi(t *testing.T) {
	result, fingerprint := IsSQLi("-1' and 1=1 union/* foo */select load_file('/etc/passwd')--")
	fmt.Println("=========result==========: ", result)
	fmt.Println("=======fingerprint=======: ", string(fingerprint))
}

type sqliTest struct {
	name        string
	input       string
	fingerprint string
}

func readTestData(t testing.TB, filename string) sqliTest {
	t.Helper()

	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	state := ""
	test := sqliTest{
		name: filename,
	}

	br := bufio.NewReaderSize(f, 8192)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		str := string(bytes.TrimSpace(line))
		if str == "--TEST--" || str == "--INPUT--" || str == "--EXPECTED--" {
			state = str
		} else {
			switch state {
			case "--INPUT--":
				test.input += str
			case "--EXPECTED--":
				test.fingerprint += str
			}
		}
	}
	test.input = strings.TrimSpace(test.input)
	test.fingerprint = strings.TrimSpace(test.fingerprint)
	return test
}

//go:embed tests
var tests embed.FS

func TestSQLiDriver(t *testing.T) {
	files, err := fs.Glob(tests, "tests/*-sqli-*")
	if err != nil {
		t.Fatal(err)
	}

	var tests []sqliTest
	for _, file := range files {
		tests = append(tests, readTestData(t, file))
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			res, fp := IsSQLi(tt.input)
			if len(tt.fingerprint) == 0 && res {
				t.Errorf("expected not sql injection but was injection")
			}
			if have, want := fp, tt.fingerprint; have != want {
				t.Errorf("incorrect fingerprint: have %s, want %s", have, want)
			}
		})
	}
}

func BenchmarkSQLiDriver(b *testing.B) {
	files, err := fs.Glob(tests, "tests/*-sqli-*")
	if err != nil {
		b.Fatal(err)
	}

	var tests []sqliTest
	for _, file := range files {
		tests = append(tests, readTestData(b, file))
	}

	for _, tc := range tests {
		tt := tc
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				res, _ := IsSQLiBenchmark(tt.input)
				if len(tt.fingerprint) == 0 && res {
					b.Errorf("expected not sql injection but was injection")
				}
			}
		})
	}
}
