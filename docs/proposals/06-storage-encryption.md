---
title: OEP Title

oep-number: 6

creation-date: 2023-01-03

status: implementable

authors:
- @Ashughorla
- @kasabe28
- @ushabelgur
  
reviewers:
- @manuel

---

# OEP-6: Storage Encryption

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)

## Summary
One of the important feature of Cloud Native IaS is to provide secure storage. This functionality is built on top of Ceph-CSI supported individual RBD PersistentVolumeClaim encryption, more information can be found here https://rook.io/docs/rook/v1.9/ceph-csi-drivers.html#enable-rbd-encryption-support . This proposal focuses on providing option to enable encryption for individual Volume.

## Motivation
As part of Storage encryption feature Onmetal API to support option to enable encryption of Volumes.
Each volume can set encryption enabled flag and pass encryption key(Optional). If encryption key is not passed, default eccrypted storage class will be used.

### Goals
  - Volume should allow to specify encryption enabled or not
  - Volume should allow to specify encryption key to be be used
  - Validation for encryptionPassphrase provided by user

### Non-Goals
  - CephCSI can generate unique passphrase (DEK Data-Encryption-Key) for each volume to be used to encrypt/decrypt data. The passphrase (DEK) is encrypted by encryptionPassphrase (KEK Key-Encryption-Key) and stored in the image metadata of the volume.
  - Encryption KMS to be used ? Like HashiCorp Vault -> https://github.com/ceph/ceph-csi/blob/v3.6.0/docs/deploy-rbd.md#encryption-for-rbd-volumes
    Or we have to consider simple passpharse ?

## Proposal
The proposal to provide storage encryption introduces new fields `encryption.enabled` and `encryption.secretRef.name` in existing `Volume` type. `encryption.enabled` is boolean field indicating whether encryption to be enabled or not for the `Volume`. `ecnryption.secretRef.name` is an optional secret for specifying encryptionPassphrase for storage class.

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