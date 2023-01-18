---
title: Storage Encryption

oep-number: 6

creation-date: 2023-01-03

status: implementable

authors:
- "@Ashughorla"
- "@kasabe28"
- "@ushabelgur"
  
reviewers:
- "@ManuStoessel"
- "@adracus"
- "@afritzler"
- "@lukasfrank"

---

# OEP-6: Storage Encryption

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Details](#Details)
- [Proposal](#proposal)

## Summary
One of the important feature of Cloud Native IaaS is to provide secure storage. This proposal focuses on providing option to enable encryption for individual onmetal Volume.

## Motivation
As part of Storage encryption feature Onmetal API to support option to enable encryption of Volumes. Each `Volume`  provide encryption key secret reference(Optional).

### Goals
  - Allow user to enable/disable volume encryption by providing secret reference
  - Encrypt volume with user provided encryption key

### Details
  - User can provide encryption key via `encryption.secretRef` to encrypt onmetal `Volume`
  - Presence of `encryption.secretRef` indicates `Volume` has to be encrypted.
  - If `encryption.secretRef` is not provided by user, then onmetal `Volume` remains unencrypted

## Proposal
The proposal to provide storage encryption introduces new field `encryption.secretRef` in existing `Volume` type. `encryption.secretRef` is an optional field for encryption key secret reference.

Volume with encryption key secret reference:

[//]: # (@formatter:off)
```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: Volume
metadata:
  name: sample-volume
  namespace: default
spec:
  volumeClassRef:
    name: fast
  volumePoolRef:
    name: ceph
  resources:
    storage: 1Gi
  encryption:
    secretRef: encryption-key-secret    # this is optional
```
[//]: # (@formatter:on)

Secret for encryption key

[//]: # (@formatter:off)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: encryption-key-secret
  namespace: default
stringData:
  encryptionKey: test-encryption
```
[//]: # (@formatter:on)