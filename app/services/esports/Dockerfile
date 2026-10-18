# Container image for the esports service.
# The build context is the workspace root (single Go module), so build from
# the repo root:  docker build -f app/services/esports/Dockerfile .
# The Nx target `esports:image` wraps this and stamps the version.

FROM golang:1.26-alpine AS build
ARG GOPROXY=https://proxy.golang.org,direct
ARG VERSION=dev
ENV CGO_ENABLED=0
WORKDIR /src
# Manifests first so `go mod download` caches independently of source edits.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-X main.Version=${VERSION}" -o /out/esports ./app/services/esports/cmd/app

FROM alpine:3.23
# busybox (wget) backs the container healthcheck; the binary is static (CGO off).
COPY --from=build /out/esports /esports
COPY app/services/esports/container/ /configs/
USER nobody
ENTRYPOINT ["/esports"]
CMD ["-conf", "/configs"]
