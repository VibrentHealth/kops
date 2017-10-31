/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awsmodel

import "k8s.io/kops/pkg/model"

type AWSModelContext struct {
	*model.KopsModelContext
}

// IAMNameForSpotFleet returns the name to use for IAM roles, policies etc for the spot fleet IAM permissions
func (b *AWSModelContext) IAMNameForSpotFleet() string {
	return "spotfleet." + b.ClusterName()
}
