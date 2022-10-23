FROM golang:1-bullseye AS builder

WORKDIR /workdir/
COPY . /workdir/

RUN make build

FROM debian:bullseye-slim

COPY --from=builder /workdir/runn ./usr/bin

ENTRYPOINT ["/entrypoint.sh"]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
