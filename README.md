# go-libinjection

go-libinjection is a library for libinjection which wraps the C [libinjection][2] library.
It provides the same API as [libinjection-go][1] and is a drop-in replacement with full
API and behavior compatibility. By default, libinjection is packaged as a WebAssembly module and accessed
with the pure Go runtime,  [wazero][3]. This means that it is compatible with any Go application, regardless
of availability of cgo.

For TinyGo applications being built for WASM, this library will perform significantly better. For Go applications,
it seems to be slower for SQL injection and faster for XSS. The API is a drop-in replacement, so it
is best to try it and benchmark to see the effect. It is likely that in the future, libinjection-go will
improve to have competitive XSS detection and this library will be relevant only to TinyGo or
case studies of wrapping vs rewriting libraries.

## Usage

go-libinjection is a standard Go library package and can be added to a go.mod file. It will work fine in
Go or TinyGo projects.

```
go get github.com/wasilibs/go-libinjection
```

Because the library is a drop-in replacement for [petar-dambovaliev/libinjection][1], an import rewrite can
make migrating code to use it simple.

```go
import "github.com/corazawaf/libinjection-go"
```

can be changed to

```go
import "github.com/wasilibs/go-libinjection"
```

### cgo

This library also supports opting into using cgo to wrap [libinjection][2] instead
of using WebAssembly. This requires having a built version of the library available -
`pkg-config` will be used to locate the library. The build tag `aho_corasick_cgo` can be used to
enable cgo support.

## Performance

Benchmarks are run against every commit in the [bench][4] workflow. GitHub action runners are highly
virtualized and do not have stable performance across runs, but the relative numbers within a run
should still be somewhat, though not precisely, informative.

### Microbenchmarks

Microbenchmarks are the same as included in libinjection-go. Full results can be
viewed in the workflow, a sample of results for one run looks like this

```
SQLiDriver/tests/test-sqli-001.txt-2                         1.27µs ± 0%      5.40µs ± 0%          2.13µs ± 2%
SQLiDriver/tests/test-sqli-002.txt-2                          838ns ± 0%      3409ns ± 1%          1441ns ± 1%
SQLiDriver/tests/test-sqli-012.txt-2                          508ns ± 1%      2039ns ± 0%           885ns ± 0%
SQLiDriver/tests/test-sqli-013.txt-2                          450ns ± 1%      1340ns ± 0%           681ns ± 1%
SQLiDriver/tests/test-sqli-014.txt-2                          412ns ± 2%      1165ns ± 0%           576ns ± 1%
SQLiDriver/tests/test-sqli-015.txt-2                         1.11µs ± 0%      4.74µs ± 0%          1.93µs ± 1%
SQLiDriver/tests/test-sqli-016.txt-2                         1.05µs ± 1%      4.57µs ± 0%          1.88µs ± 1%
SQLiDriver/tests/test-sqli-017.txt-2                         1.03µs ± 0%      4.43µs ± 0%          1.82µs ± 0%
SQLiDriver/tests/test-sqli-018.txt-2                         1.11µs ± 3%      4.98µs ± 0%          1.98µs ± 0%
SQLiDriver/tests/test-sqli-033.txt-2                          441ns ± 0%      1058ns ± 1%           566ns ± 0%
SQLiDriver/tests/test-sqli-034.txt-2                          434ns ± 1%      1482ns ± 0%           687ns ± 1%
SQLiDriver/tests/test-sqli-035.txt-2                          222ns ± 1%         3ns ± 0%             3ns ± 0%
SQLiDriver/tests/test-sqli-036.txt-2                         1.15µs ± 0%      4.48µs ± 0%          2.17µs ± 2%
SQLiDriver/tests/test-sqli-037.txt-2                         1.15µs ± 1%      4.49µs ± 0%          2.16µs ± 1%
SQLiDriver/tests/test-sqli-038.txt-2                          580ns ± 1%      2430ns ± 1%           968ns ± 1%
SQLiDriver/tests/test-sqli-049.txt-2                          883ns ± 1%      2587ns ± 0%          1427ns ± 1%
SQLiDriver/tests/test-sqli-050.txt-2                          443ns ± 2%       930ns ± 0%           586ns ± 0%
IsXSS/<script>alert(1);</script>-2                           1.12µs ± 0%      0.46µs ± 1%          0.31µs ± 1%
IsXSS/><script>alert(1);</script>-2                          1.13µs ± 0%      0.49µs ± 1%          0.32µs ± 2%
IsXSS/x_><script>alert(1);</script>-2                        1.13µs ± 0%      0.49µs ± 1%          0.32µs ± 2%
IsXSS/'_><script>alert(1);</script>-2                        1.13µs ± 0%      0.49µs ± 1%          0.32µs ± 1%
IsXSS/"><script>alert(1);</script>-2                         1.13µs ± 0%      0.49µs ± 2%          0.32µs ± 1%
IsXSS/<a_href='javascript:alert(1)'>-2                       2.08µs ± 1%      0.79µs ± 3%          0.38µs ± 1%
IsXSS/<a_href=javascript:alert(1)>-2                         2.11µs ± 1%      0.77µs ± 3%          0.51µs ± 0%
IsXSS/<a_href__=___javascript:alert(1);_>-2                  2.18µs ± 1%      0.81µs ± 2%          0.54µs ± 1%
IsXSS/<a_href="__javascript:alert(1);"_>-2                   2.14µs ± 0%      0.82µs ± 1%          0.38µs ± 0%
IsXSS/<a_href="JAVASCRIPT:alert(1);"_>-2                     2.09µs ± 0%      0.78µs ± 1%          0.37µs ± 0%
```

The library seems to consistently perform 4x slower for SQL injection, with one odd case, and 2x faster for XSS.

### wafbench

wafbench tests the performance of replacing the detect operators of the OWASP [CoreRuleSet][5] and
[Coraza][6] implementation with this library. This benchmark is considered a real world performance
test, though the bulk of processing is in regex, not detection.\

```
WAF/FTW-2                          35.6s ± 0%          36.5s ± 1%              35.3s ± 0%
WAF/POST/1-2                      4.13ms ± 2%         4.22ms ± 1%             4.10ms ± 1%
WAF/POST/1000-2                   21.6ms ± 1%         22.1ms ± 3%             21.2ms ± 1%
WAF/POST/10000-2                   189ms ± 1%          195ms ± 0%              186ms ± 1%
WAF/POST/100000-2                  1.86s ± 1%          1.93s ± 1%              1.83s ± 1%
```

Using the WebAssembly wrapped version of this library, performance is a little slower,
especially at larger payload sizes. With cgo, performance is a little faster.

[1]: https://github.com/corazawaf/libinjection-go
[2]: https://github.com/libinjection/libinjection
[3]: https://wazero.io
[4]: https://github.com/wasilibs/go-libinjection/actions/workflows/bench.yaml
[5]: https://github.com/coreruleset/coreruleset
[6]: https://github.com/corazawaf/coraza
