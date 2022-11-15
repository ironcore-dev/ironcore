#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPO_ROOT="$SCRIPT_DIR/.."
CODEGEN_PKG="${CODEGEN_PKG:-"$( (go mod download > /dev/null 2>&1) && go list -m -f '{{.Dir}}' k8s.io/code-generator)"}"

VGOPATH="$(mktemp -d)"
#trap 'rm -rf "$VGOPATH"' EXIT

# Setup virtual GOPATH so the codegen tools work as expected.
(
cd "$REPO_ROOT"
go run github.com/onmetal/vgopath "$VGOPATH"
)

export GOPATH="$VGOPATH"
export GO111MODULE=off

(
cd "$REPO_ROOT"
protoc \
  --proto_path ./ori/apis/compute/v1alpha1 \
  --proto_path "$VGOPATH/src" \
  --gogo_out=plugins=grpc:"$REPO_ROOT" \
  ./ori/apis/compute/v1alpha1/api.proto
)
