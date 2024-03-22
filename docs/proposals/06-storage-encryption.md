---
title: Storage Encryption

iep-number: 6

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

# IEP-6: Storage Encryption

## Table of Contents

- [IEP-6: Storage Encryption](#IEP-6-storage-encryption)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)

## Summary
One of the important feature of Cloud Native IaaS is to provide secure storage. This proposal focuses on providing option to enable encryption for individual ironcore Volume.

## Motivation
As part of Storage encryption feature the IronCore API supports option to enable encryption of Volumes. Volume level encryption helps protect users from data theft or accidental loss, by rendering data stored on hard drives unreadable when an unauthorized user tries to gain access. The loss of encryption keys is a major concern, as it can render any encrypted data useless. 

### Goals
  - Allow user to enable volume encryption by providing encryption key via secret reference

### Non-Goals
  - Add a new attribute to provide source of encryption key like None/UserProvidedKey/DefaultMasterKey
  - Add KMS support to manage user provided encryption keys

## Proposal
 - The proposal introduces a new field `encryption` with currently the single attribute `secretRef`, referencing a secret to use for encryption, in existing `Volume` type. 
 - `encryption` is an optional field.
 - If `encryption` field is not provided by user, then ironcore `Volume` remains unencrypted
 - To encrypt ironcore `Volume`, user has to first create kubernetes secret of Opaque type with key-value pair as below:
    - key = `encryptionKey` 
    - value = base64-encoded 256 bit encryption key
 - Then provide this secret name to `encryption.secretRef` attribute of `Volume` type.

Secret for encryption key

[//]: # (@formatter:off)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: encryption-key-secret
  namespace: default
type: Opaque
data:
  encryptionKey: QW9zejI4Y0xIR3pjR3M2UGltdHZVSnVSSGt6aWZiVTU4V3NIZElIL09idz0=
```
[//]: # (@formatter:on)

Volume with encryption key secret reference:

[//]: # (@formatter:off)
```yaml
apiVersion: storage.ironcore.dev/v1alpha1
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
    secretRef: encryption-key-secret
```
[//]: # (@formatter:on)
