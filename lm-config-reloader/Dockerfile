# Build the reloader binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/
COPY internal/ internal/

ARG VERSION_PKG
ARG LM_RELOADER_VERSION
ARG VERSION_DATE

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags="-X ${VERSION_PKG}.lmConfigReloader=${LM_RELOADER_VERSION} -X ${VERSION_PKG}.buildDate=${VERSION_DATE}" -a -o lmconfigreloader main.go

FROM alpine:3.15.0
LABEL org.opencontainers.image.source https://github.com/logicmonitor/lm-k8s-webhook
RUN apk --no-cache add ca-certificates
RUN addgroup -S -g 1001 lmuser && adduser -S lmuser -u 1001 -G lmuser
COPY --from=builder /workspace/lmconfigreloader /usr/local/bin/lmconfigreloader/
WORKDIR /usr/local/bin/lmconfigreloader
RUN chown -R lmuser:lmuser .
RUN chmod +x lmconfigreloader
USER lmuser
ENTRYPOINT ["./lmconfigreloader"]