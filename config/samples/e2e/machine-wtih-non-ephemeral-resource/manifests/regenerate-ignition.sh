#!/usr/bin/env bash

butane -d . ignition.yaml | \
  kubectl create secret generic ignition --from-file=ignition.yaml=/dev/stdin --dry-run=client -o yaml \
  > ignition-secret.yaml