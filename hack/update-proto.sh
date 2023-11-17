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

VIRTUAL_GOPATH="$(mktemp -d)"
trap 'rm -rf "$VIRTUAL_GOPATH"' EXIT

# Setup virtual GOPATH so the codegen tools work as expected.
(
cd "$REPO_ROOT"
"$VGOPATH" -o "$TGOPATH"
)

export GOPATH="$TGOPATH"
export GO111MODULE=off

function generate() {
  package="$1"
  (
  cd "$TGOPATH/src"
  export PATH="$PATH:$(dirname "$PROTOC_GEN_GOGO")"
  echo "Generating ${blue}$package${normal}"
  protoc \
    --proto_path "./github.com/ironcore-dev/ironcore/$package" \
    --proto_path "$TGOPATH/src" \
    --gogo_out=plugins=grpc:"$TGOPATH/src" \
    "./github.com/ironcore-dev/ironcore/$package/api.proto"
  )
}

generate "ori/apis/meta/v1alpha1"
generate "ori/apis/machine/v1alpha1"
generate "ori/apis/volume/v1alpha1"
generate "ori/apis/bucket/v1alpha1"
