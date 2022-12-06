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
COPY api/ api/
COPY apiutils/ apiutils/
COPY broker/ broker/
COPY client-go/ client-go/
COPY onmetal-apiserver/ onmetal-apiserver/
COPY onmetal-controller-manager/ onmetal-controller-manager/
COPY ori/ ori/
COPY orictl/ orictl/
COPY poollet/ poollet/
COPY testutils/ testutils/
COPY utils/ utils/

ARG TARGETOS
ARG TARGETARCH

RUN mkdir bin

FROM builder as apiserver-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/onmetal-apiserver ./onmetal-apiserver/cmd/apiserver

FROM builder as manager-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/onmetal-controller-manager ./onmetal-controller-manager/main.go

FROM builder as machinepoollet-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/machinepoollet ./poollet/machinepoollet/cmd/machinepoollet/main.go

FROM builder as machinebroker-builder

# TODO: Remove orictl-machine once debug containers are more broadly available.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/machinebroker ./broker/machinebroker/cmd/machinebroker/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/orictl-machine ./orictl/cmd/orictl-machine/main.go

FROM builder as orictl-machine-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/volumepoollet ./orictl/cmd/orictl-machine/main.go

FROM builder as volumepoollet-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/volumepoollet ./poollet/volumepoollet/cmd/volumepoollet/main.go


FROM builder as volumebroker-builder

# TODO: Remove orictl-volume once debug containers are more broadly available.
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/volumebroker ./broker/volumebroker/cmd/volumebroker/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/orictl-volume ./orictl/cmd/orictl-volume/main.go

FROM builder as orictl-volume-builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -ldflags="-s -w" -a -o bin/volumepoollet ./orictl/cmd/orictl-volume/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as manager
WORKDIR /
COPY --from=manager-builder /workspace/bin/onmetal-controller-manager .
USER 65532:65532

ENTRYPOINT ["/onmetal-controller-manager"]

FROM gcr.io/distroless/static:nonroot as apiserver
WORKDIR /
COPY --from=apiserver-builder /workspace/bin/onmetal-apiserver .
USER 65532:65532

ENTRYPOINT ["/onmetal-apiserver"]

FROM gcr.io/distroless/static:nonroot as machinepoollet
WORKDIR /
COPY --from=machinepoollet-builder /workspace/bin/machinepoollet .
USER 65532:65532

ENTRYPOINT ["/machinepoollet"]

# TODO: Switch to distroless as soon as ephemeral debug containers are more broadly available.
FROM debian:bullseye-slim as machinebroker
WORKDIR /
COPY --from=machinebroker-builder /workspace/bin/machinebroker .
# TODO: Remove orictl-machine as soon as ephemeral debug containers are more broadly available.
COPY --from=machinebroker-builder /workspace/bin/orictl-machine .
USER 65532:65532

ENTRYPOINT ["/machinebroker"]

FROM debian:bullseye-slim as orictl-machine
WORKDIR /
COPY --from=orictl-machine-builder /workspace/bin/orictl-machine .
USER 65532:65532

FROM gcr.io/distroless/static:nonroot as volumepoollet
WORKDIR /
COPY --from=volumepoollet-builder /workspace/bin/volumepoollet .
USER 65532:65532

ENTRYPOINT ["/volumepoollet"]

# TODO: Switch to distroless as soon as ephemeral debug containers are more broadly available.
FROM debian:bullseye-slim as volumebroker
WORKDIR /
COPY --from=volumebroker-builder /workspace/bin/volumebroker .
# TODO: Remove orictl-volume as soon as ephemeral debug containers are more broadly available.
COPY --from=volumebroker-builder /workspace/bin/orictl-volume .
USER 65532:65532

ENTRYPOINT ["/volumebroker"]

FROM debian:bullseye-slim as orictl-volume
WORKDIR /
COPY --from=orictl-volume-builder /workspace/bin/orictl-volume .
USER 65532:65532
