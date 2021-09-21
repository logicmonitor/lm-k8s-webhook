# Build the manager binary
FROM golang:1.16 as builder

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

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o lmwebhook main.go

FROM alpine:3.14.2
RUN apk --no-cache add ca-certificates
RUN addgroup -S -g 1001 lmuser && adduser -S lmuser -u 1001 -G lmuser
COPY --from=builder /workspace/lmwebhook /usr/local/bin/webhook/
WORKDIR /usr/local/bin/webhook
RUN chown -R lmuser:lmuser .
RUN chmod +x lmwebhook
USER lmuser
ENTRYPOINT ["./lmwebhook"]