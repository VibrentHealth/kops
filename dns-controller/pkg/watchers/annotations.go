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

package watchers

// AnnotationNameDnsExternal is used to set up a DNS name for accessing the resource from outside the cluster
// For a service of Type=LoadBalancer, it would map to the external LB hostname or IP
const AnnotationNameDnsExternal = "dns.alpha.kubernetes.io/external"

// AnnotationNameDnsInternal is used to set up a DNS name for accessing the resource from inside the cluster
// This is only supported on Pods currently, and maps to the Internal address
const AnnotationNameDnsInternal = "dns.alpha.kubernetes.io/internal"

// AnnotationNameExternalIP is used to set the IP to expose
// This is useful when the node does not have an ExternalIP configured
const AnnotationNameExternalIP = "dns.alpha.kubernetes.io/external-ip"
