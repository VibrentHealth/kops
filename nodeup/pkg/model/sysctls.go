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

package model

import (
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// SysctlBuilder set up our sysctls
type SysctlBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &SysctlBuilder{}

func (b *SysctlBuilder) Build(c *fi.ModelBuilderContext) error {
	var sysctls []string

	// Common settings
	{
		sysctls = append(sysctls,
			"# Kubernetes Settings",
			"")

		// A higher vm.max_map_count is great for elasticsearch, mongo, or other mmap users
		// See https://github.com/kubernetes/kops/issues/1340
		sysctls = append(sysctls, "vm.max_map_count = 262144",
			"")

		// See https://github.com/kubernetes/kubernetes/pull/38001
		sysctls = append(sysctls,
			"kernel.softlockup_panic = 1",
			"kernel.softlockup_all_cpu_backtrace = 1",
			"")

		// See https://github.com/kubernetes/kube-deploy/issues/261
		sysctls = append(sysctls,
			"# Increase the number of connections",
			"net.core.somaxconn = 32768",
			"",

			"# Maximum Socket Receive Buffer",
			"net.core.rmem_max = 16777216",
			"",

			"# Default Socket Send Buffer",
			"net.core.wmem_max = 16777216",
			"",

			"# Increase the maximum total buffer-space allocatable",
			"net.ipv4.tcp_wmem = 4096 12582912 16777216",
			"net.ipv4.tcp_rmem = 4096 12582912 16777216",
			"",

			"# Increase the number of outstanding syn requests allowed",
			"net.ipv4.tcp_max_syn_backlog = 8096",
			"",

			"# For persistent HTTP connections",
			"net.ipv4.tcp_slow_start_after_idle = 0",
			"",

			"# Increase the tcp-time-wait buckets pool size to prevent simple DOS attacks",
			"net.ipv4.tcp_tw_reuse = 1",
			"",

			// We can't change the local_port_range without changing the NodePort range
			//"# Allowed local port range",
			//"net.ipv4.ip_local_port_range = 10240 65535",
			//"",

			"# Max number of packets that can be queued on interface input",
			"# If kernel is receiving packets faster than can be processed",
			"# this queue increases",
			"net.core.netdev_max_backlog = 16384",
			"",

			"# Increase size of file handles and inode cache",
			"fs.file-max = 2097152",
			"",

			"# Increase size of conntrack table size to avoid poor iptables performance",
			"net.netfilter.nf_conntrack_max = 1000000",
			"",
		)
	}

	if b.Cluster.Spec.CloudProvider == string(fi.CloudProviderAWS) {
		sysctls = append(sysctls,
			"# AWS settings",
			"",
			"# Issue #23395",
			"net.ipv4.neigh.default.gc_thresh1=0",
			"")
	}

	if b.Cluster.Spec.CloudProvider == string(fi.CloudProviderGCE) {
		sysctls = append(sysctls,
			"# GCE settings",
			"",
			"net.ipv4.ip_forward=1",
			"")
	}

	t := &nodetasks.File{
		Path:            "/etc/sysctl.d/99-k8s-general.conf",
		Contents:        fi.NewStringResource(strings.Join(sysctls, "\n")),
		Type:            nodetasks.FileType_File,
		OnChangeExecute: [][]string{{"sysctl", "--system"}},
	}
	c.AddTask(t)

	return nil
}
