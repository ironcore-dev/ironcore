#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPO_ROOT="$SCRIPT_DIR/.."
export TERM="xterm-256color"

blue="$(tput setaf 4)"
normal="$(tput sgr0)"

VGOPATH="$VGOPATH"
PROTOC_GEN_GOGO="$PROTOC_GEN_GOGO"

TGOPATH="$(mktemp -d)"
trap 'rm -rf "$TGOPATH"' EXIT

# Setup virtual GOPATH so the codegen tools work as expected.
(
cd "$REPO_ROOT"
"$VGOPATH" "$TGOPATH"
)

export GOPATH="$TGOPATH"
export GO111MODULE=off

(
cd "$REPO_ROOT"
export PATH="$PATH:$(dirname "$PROTOC_GEN_GOGO")"
echo "Generating ${blue}ori/machine${normal}"
protoc \
  --proto_path ./ori/apis/machine/v1alpha1 \
  --proto_path "$TGOPATH/src" \
  --gogo_out=plugins=grpc:"$REPO_ROOT" \
  ./ori/apis/machine/v1alpha1/api.proto
)

(
cd "$REPO_ROOT"
export PATH="$PATH:$(dirname "$PROTOC_GEN_GOGO")"
echo "Generating ${blue}ori/volume${normal}"
protoc \
  --proto_path ./ori/apis/volume/v1alpha1 \
  --proto_path "$TGOPATH/src" \
  --gogo_out=plugins=grpc:"$REPO_ROOT" \
  ./ori/apis/volume/v1alpha1/api.proto
)
