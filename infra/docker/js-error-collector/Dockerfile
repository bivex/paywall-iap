FROM alpine:3.20 AS builder

RUN apk add --no-cache build-base

WORKDIR /src
COPY tools/js-error-collector/main.cpp ./main.cpp
RUN g++ -std=c++17 -O2 -static -s -o /out/js-error-collector main.cpp

FROM scratch

ENV PORT=8080 \
    LOG_PATH=/data/frontend-errors.ndjson \
    MAX_BODY_BYTES=65536

COPY --from=builder /out/js-error-collector /js-error-collector

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["/js-error-collector"]