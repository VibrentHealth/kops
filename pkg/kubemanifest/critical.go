/*
Copyright 2017 The Kubernetes Authors.

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

package kubemanifest

import (
	"k8s.io/api/core/v1"
)

// MarkPodAsCritical adds the required annotations for a pod to be considered critical
func MarkPodAsCritical(pod *v1.Pod) {
	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = make(map[string]string)
	}
	pod.ObjectMeta.Annotations["scheduler.alpha.kubernetes.io/critical-pod"] = ""

	toleration := v1.Toleration{
		Key:      "CriticalAddonsOnly",
		Operator: v1.TolerationOpExists,
	}
	pod.Spec.Tolerations = append(pod.Spec.Tolerations, toleration)
}

// MapEtcdHosts map the /etc/hosts file into the pod (useful for gossip DNS)
func MapEtcdHosts(pod *v1.Pod, container *v1.Container) {
	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:      "hosts",
		MountPath: "/etc/hosts",
		ReadOnly:  true,
	})
	hostPathFile := v1.HostPathFile
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "hosts",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/etc/hosts",
				Type: &hostPathFile,
			},
		},
	})
}
