# Build the manager binary
FROM golang:1.13-alpine as builder

WORKDIR /workspace

ADD https://storage.googleapis.com/kubernetes-release/release/v1.6.4/bin/linux/amd64/kubectl /workspace/kubectl
RUN chmod +x /workspace/kubectl

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o shalm main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

WORKDIR /app
ENV HOME=/app
COPY --from=builder /workspace/shalm .
COPY --from=builder /workspace/kubectl /usr/bin/kubectl
USER nonroot:nonroot

ENTRYPOINT ["/app/shalm","controller"]
