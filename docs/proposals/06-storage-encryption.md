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
As part of Storage encryption feature Onmetal API to support option to enable encryption of Volumes. Each `Volume` can set encryption enabled flag and provide encryption key secret reference(Optional).

### Goals
  - Allow user to enable/disable volume encryption with flag
  - Encrypt volume with user provided encryption key

### Details
  - By default encryption enabled flag will be set to false
  - If encryption enabled flag is set to true, then user can provide encryption key via secret reference to encrypt onmetal `Volume`.
  - If encryption enabled flag is set to true and user has not provided encryption key secret reference in `Volume`, then `Volume` will be encrypted with shared key of storage provider.
  - User can provide `volumePoolSelector` label to look for encrypted `VolumePool`

## Proposal
The proposal to provide storage encryption introduces new field `encryption.enabled` and `encryption.secretRef` in existing `Volume` type. `encryption.enabled` is boolean field indicating whether encryption to be enabled or not for the `Volume`. `encryption.secretRef` is an optional field for encryption key secret reference.

Volume with encryption enabled flag:

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
    enabled: true
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