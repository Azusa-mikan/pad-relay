FROM ubuntu:26.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        gcc \
        golang-go \
        libsdl2-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# Cache Go modules in the builder image; source files are mounted at runtime.
COPY go.mod go.sum ./
RUN go mod download

COPY build/script/padrelay-static-build.sh /usr/local/bin/padrelay-static-build
COPY build/script/padrelay-static-pkg-config /usr/local/bin/padrelay-static-pkg-config
RUN chmod +x /usr/local/bin/padrelay-static-build \
    /usr/local/bin/padrelay-static-pkg-config

ENTRYPOINT ["/usr/local/bin/padrelay-static-build"]
