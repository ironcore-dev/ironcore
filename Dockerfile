# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY api/ api/
COPY broker/ broker/
COPY client-go/ client-go/
COPY cmd/ cmd/
COPY internal/ internal/
COPY iri/ iri/
COPY irictl/ irictl/
COPY irictl-bucket/ irictl-bucket/
COPY irictl-machine/ irictl-machine/
COPY irictl-volume/ irictl-volume/
COPY poollet/ poollet/
COPY utils/ utils/

ARG TARGETOS
ARG TARGETARCH

RUN mkdir bin

FROM builder AS apiserver-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/ironcore-apiserver ./cmd/ironcore-apiserver

FROM builder AS manager-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/ironcore-controller-manager ./cmd/ironcore-controller-manager

FROM builder AS machinepoollet-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/machinepoollet ./poollet/machinepoollet/cmd/machinepoollet/main.go

FROM builder AS machinebroker-builder

# TODO: Remove irictl-machine once debug containers are more broadly available.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/machinebroker ./broker/machinebroker/cmd/machinebroker/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/irictl-machine ./irictl-machine/cmd/irictl-machine/main.go

FROM builder AS irictl-machine-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/irictl-machine ./irictl-machine/cmd/irictl-machine/main.go

FROM builder AS volumepoollet-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/volumepoollet ./poollet/volumepoollet/cmd/volumepoollet/main.go


FROM builder AS volumebroker-builder

# TODO: Remove irictl-volume once debug containers are more broadly available.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/volumebroker ./broker/volumebroker/cmd/volumebroker/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/irictl-volume ./irictl-volume/cmd/irictl-volume/main.go

FROM builder AS irictl-volume-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/irictl-volume ./irictl-volume/cmd/irictl-volume/main.go

FROM builder AS bucketpoollet-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/bucketpoollet ./poollet/bucketpoollet/cmd/bucketpoollet/main.go


FROM builder AS bucketbroker-builder

# TODO: Remove irictl-bucket once debug containers are more broadly available.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/bucketbroker ./broker/bucketbroker/cmd/bucketbroker/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/irictl-bucket ./irictl-bucket/cmd/irictl-bucket/main.go

FROM builder AS irictl-bucket-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/irictl-bucket ./irictl-bucket/cmd/irictl-bucket/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS manager
WORKDIR /
COPY --from=manager-builder /workspace/bin/ironcore-controller-manager .
USER 65532:65532

ENTRYPOINT ["/ironcore-controller-manager"]

FROM gcr.io/distroless/static:nonroot AS apiserver
WORKDIR /
COPY --from=apiserver-builder /workspace/bin/ironcore-apiserver .
USER 65532:65532

ENTRYPOINT ["/ironcore-apiserver"]

FROM gcr.io/distroless/static:nonroot AS machinepoollet
WORKDIR /
COPY --from=machinepoollet-builder /workspace/bin/machinepoollet .
USER 65532:65532

ENTRYPOINT ["/machinepoollet"]

# TODO: Switch to distroless as soon as ephemeral debug containers are more broadly available.
FROM debian:bullseye-slim AS machinebroker
WORKDIR /
COPY --from=machinebroker-builder /workspace/bin/machinebroker .
# TODO: Remove irictl-machine as soon as ephemeral debug containers are more broadly available.
COPY --from=machinebroker-builder /workspace/bin/irictl-machine .
USER 65532:65532

ENTRYPOINT ["/machinebroker"]

FROM debian:bullseye-slim AS irictl-machine
WORKDIR /
COPY --from=irictl-machine-builder /workspace/bin/irictl-machine .
USER 65532:65532

FROM gcr.io/distroless/static:nonroot AS volumepoollet
WORKDIR /
COPY --from=volumepoollet-builder /workspace/bin/volumepoollet .
USER 65532:65532

ENTRYPOINT ["/volumepoollet"]

# TODO: Switch to distroless as soon as ephemeral debug containers are more broadly available.
FROM debian:bullseye-slim AS volumebroker
WORKDIR /
COPY --from=volumebroker-builder /workspace/bin/volumebroker .
# TODO: Remove irictl-volume as soon as ephemeral debug containers are more broadly available.
COPY --from=volumebroker-builder /workspace/bin/irictl-volume .
USER 65532:65532

ENTRYPOINT ["/volumebroker"]

FROM debian:bullseye-slim AS irictl-volume
WORKDIR /
COPY --from=irictl-volume-builder /workspace/bin/irictl-volume .
USER 65532:65532

FROM gcr.io/distroless/static:nonroot AS bucketpoollet
WORKDIR /
COPY --from=bucketpoollet-builder /workspace/bin/bucketpoollet .
USER 65532:65532

ENTRYPOINT ["/bucketpoollet"]

# TODO: Switch to distroless as soon as ephemeral debug containers are more broadly available.
FROM debian:bullseye-slim AS bucketbroker
WORKDIR /
COPY --from=bucketbroker-builder /workspace/bin/bucketbroker .
# TODO: Remove irictl-bucket as soon as ephemeral debug containers are more broadly available.
COPY --from=bucketbroker-builder /workspace/bin/irictl-bucket .
USER 65532:65532

ENTRYPOINT ["/bucketbroker"]

FROM debian:bullseye-slim AS irictl-bucket
WORKDIR /
COPY --from=irictl-bucket-builder /workspace/bin/irictl-bucket .
USER 65532:65532
