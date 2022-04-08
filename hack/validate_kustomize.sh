#!/usr/bin/env bash

set -e

BASEDIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

for dir in "$BASEDIR"/../config/**/*; do
  [[ -e "$dir" ]] || break
  [[ -f "$dir/kustomization.yaml" ]] || break
  [[ "$dir" != *"config/samples"* ]] || break
  echo "Validating $dir"
  kustomize build "$dir" > /dev/null
done
