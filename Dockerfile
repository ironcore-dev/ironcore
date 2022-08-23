# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.19 as builder

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
COPY apis/ apis/
COPY apiserver/ apiserver/
COPY app/ app/
COPY cmd/ cmd/
COPY clientutils/ clientutils/
COPY controllers/ controllers/
COPY equality/ equality/
COPY generated/ generated/
COPY machinepoollet/ machinepoollet/
COPY registry/ registry/
COPY serializer/ serializer/
COPY tableconvertor/ tableconvertor/

ARG TARGETOS TARGETARCH

RUN mkdir bin

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/manager main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/apiserver ./cmd/apiserver

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as manager
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

FROM gcr.io/distroless/static:nonroot as apiserver
WORKDIR /
COPY --from=builder /workspace/bin/apiserver .
USER 65532:65532

ENTRYPOINT ["/apiserver"]
