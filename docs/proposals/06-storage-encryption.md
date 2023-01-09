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
    - [Details](#Details)
- [Proposal](#proposal)

## Summary
One of the important feature of Cloud Native IaS is to provide secure storage. This functionality is built on top of Ceph-CSI supported individual RBD PersistentVolumeClaim encryption, more information can be found here https://rook.io/docs/rook/v1.9/ceph-csi-drivers.html#enable-rbd-encryption-support. This proposal focuses on providing option to enable encryption for individual Volume.

## Motivation
As part of Storage encryption feature Onmetal API to support option to enable encryption of Volumes.
Each `VolumePool` can set encryption enabled flag and provide secret reference holding encryption key(Optional). If encryption key is not passed, then encryption key is fetched from storage class secrets.

### Goals
  - VolumePool should provide encryption enabled flag
  - VolumePool should provide secret name holding `encryptionPassphrase`
  - Volume should update staus with encrypted flag

### Details
  - By default encryption enabled flag and encrypted flag will be set to false
  - Generate random string corresponding to `encryptionPassphrase` key using `math/rand` package. An `encryptionPassphrase` should have at least 20 characters and should be difficult to guess. It should contain upper case letters, lower case letters, digits and at least one punctuation character.
  - Create secret holding `encryptionPassphrase` as key and generated random string as value.
  - Provide secret name in VolumePool configuration and pass secret towards Storage Cluster.
  - Once the `Volume` gets encrypted with provided secret, update `Volume.status` with `encrypted: true`.

## Proposal
The proposal to provide storage encryption introduces new fields `encryption.enabled`, `encryption.secretRef.name` in existing `VolumePool` type and `status.encrypted` in existing `Volume` type. `encryption.enabled` is boolean field indicating whether encryption to be enabled or not for the `Volume`. `encryption.secretRef.name` is an secret for specifying `encryptionPassphrase` for storage class. `status.encrypted` is boolean field indicating whether `Volume` is encrypted or not.

VolumePool with encryption configuration:

[//]: # (@formatter:off)
```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: VolumePool
metadata:
  name: volumepool-sample
spec:
  providerID: onmetal://shared
  encryption:
    enabled: true
    secretRef:
      name: storage-encryption-secret
```
[//]: # (@formatter:on)

Secret for encryption passphrase:

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

Volume with status encrypted:

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
status:
  encrypted: true
```
[//]: # (@formatter:on)