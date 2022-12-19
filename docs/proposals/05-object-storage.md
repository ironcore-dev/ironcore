---
title: Object Storage

oep-number: 5

creation-date: 2022-12-19

status: implementable

authors:

- @lukasfrank
- @gehoern

  reviewers:

- @adracus
- @MalteJ

---

# OEP-5: Object Storage

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary
Object storage builds the basis for many cloud applications. An Object Storage provides a simplified object 
model for files, but has a reduced set of security and access features (non-posix). This functionality is build on top 
of the HTTP protocol. The current market standard is [S3](https://docs.aws.amazon.com/AmazonS3/latest/API/Type_API_Reference.html). 
This document describes how to integrate simplified S3 buckets into the OnMetal API without taking too many details of 
the S3 feature completeness itself.

## Motivation
Object Storage is demanded by cloud native applications, therefore, OnMetal needs to provide it for a complete solution.
The Object Storage service should be integrated into OnMetal and use all available optimizations. It is not designed 
to be _just_ a service on top of OnMetal. The used protocol is called [S3](https://docs.aws.amazon.com/AmazonS3/latest/API/Type_API_Reference.html)
which introduces a storage entity called bucket. For the beginning only the bucket creation and removal is covered.

### Goals
- S3-compatible object storage implementation
- Automatic assigned public storage endpoint (addressed via DNS)
- Providing a REST-API endpoint to address the bucket

### Non-Goals
- Support other object storage protocols than S3
- Internal buckets (reachable only from inside the cluster)
- Quota handling (size or object number limitation)
- External Object level access control (beyond what the S3 implementation provides)

## Proposal
The proposal to provide an Object Storage consists of three API resources: `Bucket`, `BucketClass` and `BucketPool`. 
A `Bucket` is the S3 enabled storage endpoint. IOPS/Bandwidth limitations are controlled via a `BucketClass` and the 
capabilities of the underlying storage provider are expressed via a `BucketPool`. A `Bucket` can be requested 
from a `BucketPool` as long as it can provide the performance characteristics described in the `BucketClass`. 
The proposed API resources are similar to `Volume`, `VolumeClass` and `Volumepool` except that a volume is a 
block device with a specific driver.


### Bucket

A `Bucket` is a *namespaced* resource to request S3-compatible object storage. 
`Bucket`s with the `type` `Public` set (only valid type for now) are accessible from the internet. 
In the future, the `Bucket`s can be extended with a `networkRef` to support other types. 
The desired `BucketClass` is referenced by the `bucketClassRef`. If no pool is pre-defined, 
the `bucketPoolSelector` will be used to find a suitable `BucketPool`.  The desired pool, either pre-defined or 
set by another controller, is stated in the `bucketPoolRef`.

The information to access the requested `Bucket` is in the `access` field of the status. 
The `endpoint` defines the address of the `Bucket` Rest-API. Access credentials are placed in a secret with is referenced 
through the `secretRef`. The `state` indicates if the `Bucket` is `Available`, `Pending` or in an `Error` state.

[//]: # (@formatter:off)
```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: Bucket
metadata:
  name: bucket-1
spec:
  # currently only type public is defined
  type: Public
  bucketClassRef:
    name: slow
  bucketPoolSelector:
    matchLabels:
      key: db
      foo: bar
  bucketPoolRef:
    name: fra-shared
status:
  access:
    endpoint: foo.bar.example.org 
    secretRef:
      name: 000225194345f27a40257c5777c96a03ce219f96731f22afc45b7dfda7d077d
  state: Available
```
[//]: # (@formatter:on)

### BucketClass

A `BucketClass` is a *non-namespaced* resource which describes the characteristics of a `Bucket`. The maximal 
performance the `Bucket` will offer (like I/O operations or throughput) is defined in the `capabilities` field.

[//]: # (@formatter:off)
```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: BucketClass
metadata:
  name: slow
capabilities:
  iops: 10
  tps: 20Mi
```
[//]: # (@formatter:on)

### BucketPool

A `BucketPool` is a *non-namespaced* logical unit and accommodates a collection of `Bucket`s. 
The `provider`'s id (the implementor's id) of the `BucketPool` is stated in the `providerID` field. 
Only `Bucket`s who tolerate all the taints, will land in the `BucketPool`. `BucketClasses` which can be fulfilled by 
the provider of the`BucketPool`, are listed in the status field `availableBucketClasses`. The `state` in the status 
indicates if the pool is `Available`, `Pending` or `Unavailable`.

[//]: # (@formatter:off)
```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: BucketPool
metadata:
  name: ceph-object-store
spec:
  providerID: cephlet://pool
  taints: []
status:
  availableBucketClasses:
    - name: fast
    - name: slow
  state: Available
```
[//]: # (@formatter:on)

## Alternatives