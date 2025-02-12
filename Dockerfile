FROM golang:1-bullseye AS builder

WORKDIR /workdir/
COPY . /workdir/

RUN apt-get update

RUN update-ca-certificates

RUN make build

FROM debian:bullseye-slim

RUN apt-get update \
    && dpkg --configure -a \
    && apt-get install -y fonts-noto-cjk \
    && apt-get install -y chromium \
    && apt-get install -y git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workdir/runn ./usr/bin

ENTRYPOINT ["/entrypoint.sh"]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
