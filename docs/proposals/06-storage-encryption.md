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

---

# OEP-6: Storage Encryption

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)

## Summary
One of the important feature of Cloud Native IaS is to provide secure storage. This functionality is built on top of Ceph-CSI supported individual RBD PersistentVolumeClaim encryption, more information can be found here https://rook.io/docs/rook/v1.9/ceph-csi-drivers.html#enable-rbd-encryption-support. This proposal focuses on providing option to enable encryption for individual Volume.

## Motivation
As part of Storage encryption feature Onmetal API to support option to enable encryption of Volumes.
Each `Volume` can set encryption enabled flag and provide secret reference holding encryption key(Optional). If encryption key is not passed, then encryption key is fetched from storage class secrets.

### Goals
  - Volume should provide encryption enabled flag
  - Volume should provide secret name holding `encryptionPassphrase`

### Details
  - By default encryption enabled flag will be set to false
  - Generate random string corresponding to `encryptionPassphrase` key using `encoding/base64` golang package
  - Create secret holding `encryptionPassphrase` as key and generated random string as value.
  - Provide secret name for Volume encryption.

## Proposal
The proposal to provide storage encryption introduces new fields `encryption.enabled` and `encryption.secretRef.name` in existing `Volume` type. `encryption.enabled` is boolean field indicating whether encryption to be enabled or not for the `Volume`. `ecnryption.secretRef.name` is an secret for specifying `encryptionPassphrase` for storage class.

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
    secretRef:
      name: storage-encryption-secret
```
[//]: # (@formatter:on)

Secret for passphrase

[//]: # (@formatter:off)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: storage-encryption-secret
  namespace: rook-ceph
stringData:
  encryptionPassphrase: test-encryption
  ```
[//]: # (@formatter:on)