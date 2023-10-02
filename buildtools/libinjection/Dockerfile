# Copyright 2022 The OWASP Coraza contributors
# SPDX-License-Identifier: Apache-2.0

FROM ghcr.io/webassembly/wasi-sdk:wasi-sdk-20

ENV CFLAGS -O3 ${CFLAGS}
ENV LDFLAGS $CFLAGS -mexec-model=reactor

RUN apt-get update && apt-get install -y autogen autoconf automake libtool binaryen curl

RUN mkdir -p /libinjection && curl -L https://github.com/libinjection/libinjection/archive/4aa3894b21d03d9d8fc364505c0617d2aca73fc1.tar.gz | tar -xz --strip-components 1 -C /libinjection
WORKDIR /libinjection
RUN ./autogen.sh && \
    # host is required by configure but not used so set it arbitrarily
    ./configure --host=i686-pc-linux-gnu --enable-shared=false && \
    make -C src libinjection.la

# Separate step so exports can be updated without recompiling.
# Number of layers isn't really a concern for this image.
# global-base=1024 same as emcc and allows further wasm-opt optimizations
RUN $CXX -o src/.libs/libinjection-noopt.so -Wl,--global-base=1024 $LDFLAGS \
    src/.libs/libinjection.a \
    -Wl,--export=malloc \
    -Wl,--export=free \
    -Wl,--export=libinjection_sqli \
    -Wl,--export=libinjection_xss

RUN wasm-opt -o src/.libs/libinjection.so --low-memory-unused --flatten --rereloop --converge -O3 src/.libs/libinjection-noopt.so

CMD ["cp", "src/.libs/libinjection.a", "src/.libs/libinjection.so", "/out/"]
