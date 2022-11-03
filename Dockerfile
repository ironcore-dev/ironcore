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
COPY apis/ apis/
COPY apiutils/ apiutils/
COPY generated/ generated/
COPY machinepoollet/ machinepoollet/
COPY onmetal-apiserver/ onmetal-apiserver/
COPY onmetal-controller-manager/ onmetal-controller-manager/
COPY ori/ ori/
COPY testutils/ testutils/
COPY utils/ utils/

ARG TARGETOS TARGETARCH

RUN mkdir bin

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/onmetal-controller-manager ./onmetal-controller-manager/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/onmetal-apiserver ./onmetal-apiserver/cmd/apiserver

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as manager
WORKDIR /
COPY --from=builder /workspace/bin/onmetal-controller-manager .
USER 65532:65532

ENTRYPOINT ["/onmetal-controller-manager"]

FROM gcr.io/distroless/static:nonroot as apiserver
WORKDIR /
COPY --from=builder /workspace/bin/onmetal-apiserver .
USER 65532:65532

ENTRYPOINT ["/onmetal-apiserver"]
