// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

const (
	BucketUIDLabel       = "bucketpoollet.ironcore.dev/bucket-uid"
	BucketNamespaceLabel = "bucketpoollet.ironcore.dev/bucket-namespace"
	BucketNameLabel      = "bucketpoollet.ironcore.dev/bucket-name"

	FieldOwner      = "bucketpoollet.ironcore.dev/field-owner"
	BucketFinalizer = "bucketpoollet.ironcore.dev/bucket"

	// DownwardAPIPrefix is the prefix for any downward label.
	BucketDownwardAPIPrefix = "downward-api.bucketpoollet.ironcore.dev/"
)
