# Copyright 2022 The OWASP Coraza contributors
# SPDX-License-Identifier: Apache-2.0

FROM ghcr.io/webassembly/wasi-sdk:wasi-sdk-34 AS wasm

ENV CFLAGS -O3 ${CFLAGS}
ENV LDFLAGS $CFLAGS -mexec-model=reactor

RUN apt-get update && apt-get install -y autogen autoconf automake libtool binaryen curl

COPY buildtools/wasm/version.txt /version.txt
RUN mkdir -p /libinjection && curl -L https://github.com/libinjection/libinjection/archive/$(cat /version.txt | awk '{$1=$1};1').tar.gz | tar -xz --strip-components 1 -C /libinjection
WORKDIR /libinjection
# host is required by configure but not used so set it arbitrarily.
RUN ./autogen.sh \
  && ./configure --host=i686-pc-linux-gnu --enable-shared=false \
  && make -C src libinjection.la

# Separate step so exports can be updated without recompiling.
# Number of layers isn't really a concern for this image.
# The linker places the stack at the start of memory by default so that stack
# overflow traps instead of overwriting data.
RUN $CXX -o src/.libs/libinjection-noopt.wasm $LDFLAGS \
  src/.libs/libinjection.a \
  -Wl,--export=malloc \
  -Wl,--export=free \
  -Wl,--export=libinjection_sqli \
  -Wl,--export=libinjection_xss

RUN wasm-opt -o src/.libs/libinjection.wasm --flatten --rereloop --converge -O3 src/.libs/libinjection-noopt.wasm

FROM golang:1.27-trixie

COPY --from=wasm /libinjection/src/.libs/libinjection.a /build/
COPY --from=wasm /libinjection/src/.libs/libinjection.wasm /build/
WORKDIR /build
RUN go run github.com/ncruces/wasm2go@v0.4.20 -unsafe -pkg=wasm2go < libinjection.wasm > libinjection.wasm.go

CMD ["cp", "/build/libinjection.a", "/build/libinjection.wasm.go", "/out/"]
