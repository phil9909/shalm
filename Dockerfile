# Build the manager binary
FROM alpine as kubectl-builder

WORKDIR /workspace

ADD https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl /workspace/kubectl
RUN chmod +x /workspace/kubectl

FROM golang:1.13-alpine as shalm-builder

WORKDIR /workspace

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
# FROM gcr.io/distroless/static
FROM alpine

WORKDIR /app
ENV HOME=/app
COPY --from=kubectl-builder /workspace/kubectl /usr/bin/kubectl
COPY --from=shalm-builder /workspace/shalm .

ENTRYPOINT ["/app/shalm","controller"]
