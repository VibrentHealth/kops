package gce

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/golang/glog"
	context "golang.org/x/net/context"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func PerformNetworkAssignments(c *kops.Cluster, cloudObj fi.Cloud) error {
	ctx := context.TODO()

	if c.Spec.NonMasqueradeCIDR != "" && c.Spec.ServiceClusterIPRange != "" {
		return nil
	}

	networkName := c.Spec.NetworkID
	if networkName == "" {
		networkName = "default"
	}

	cloud := cloudObj.(GCECloud)

	var regions []*compute.Region
	if err := cloud.Compute().Regions.List(cloud.Project()).Pages(ctx, func(p *compute.RegionList) error {
		for _, r := range p.Items {
			regions = append(regions, r)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error listing Regions: %v", err)
	}

	network, err := cloud.Compute().Networks.Get(cloud.Project(), networkName).Do()
	if err != nil {
		return fmt.Errorf("error fetching network name %q: %v", networkName, err)
	}

	subnetURLs := make(map[string]bool)
	for _, subnet := range network.Subnetworks {
		subnetURLs[subnet] = true
	}

	glog.Infof("scanning regions for subnetwork CIDR allocations")

	var subnets []*compute.Subnetwork
	for _, r := range regions {
		if err := cloud.Compute().Subnetworks.List(cloud.Project(), r.Name).Pages(ctx, func(p *compute.SubnetworkList) error {
			for _, s := range p.Items {
				subnets = append(subnets, s)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("error listing Subnetworks: %v", err)
		}
	}

	var used cidrMap
	for _, subnet := range subnets {
		if !subnetURLs[subnet.SelfLink] {
			continue
		}
		if err := used.MarkInUse(subnet.IpCidrRange); err != nil {
			return err
		}

		for _, s := range subnet.SecondaryIpRanges {
			if err := used.MarkInUse(s.IpCidrRange); err != nil {
				return err
			}
		}
	}

	podCIDR, err := used.Allocate("10.0.0.0/8", 14)
	if err != nil {
		return err
	}

	serviceCIDR, err := used.Allocate("10.0.0.0/8", 20)
	if err != nil {
		return err
	}

	glog.Infof("Will use %s for Pods and %s for Services", podCIDR, serviceCIDR)

	c.Spec.NonMasqueradeCIDR = podCIDR
	c.Spec.ServiceClusterIPRange = serviceCIDR

	return nil
}

type cidrMap struct {
	used []net.IPNet
}

func (c *cidrMap) MarkInUse(s string) error {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("error parsing network cidr %q: %v", s, err)
	}
	c.used = append(c.used, *cidr)
	return nil
}

func (c *cidrMap) Allocate(from string, mask int) (string, error) {
	_, cidr, err := net.ParseCIDR(from)
	if err != nil {
		return "", fmt.Errorf("error parsing CIDR %q: %v", from, err)
	}

	i := *cidr
	i.Mask = net.CIDRMask(mask, 32)

	for {

		ip4 := i.IP.To4()
		if ip4 == nil {
			return "", fmt.Errorf("expected IPv4 address: %v", from)
		}

		// Note we increment first, so we won't ever use 10.0.0.0/n
		n := binary.BigEndian.Uint32(ip4)
		n += 1 << uint(32-mask)
		binary.BigEndian.PutUint32(i.IP, n)

		if !cidrsOverlap(cidr, &i) {
			break
		}

		if !c.isInUse(&i) {
			if err := c.MarkInUse(i.String()); err != nil {
				return "", err
			}
			return i.String(), nil
		}
	}

	return "", fmt.Errorf("cannot allocate CIDR of size %d", mask)
}

func (c *cidrMap) isInUse(n *net.IPNet) bool {
	for i := range c.used {
		if cidrsOverlap(&c.used[i], n) {
			return true
		}
	}
	return false
}

// cidrsOverlap returns true iff the two CIDRs are non-disjoint
func cidrsOverlap(l, r *net.IPNet) bool {
	return l.Contains(r.IP) || r.Contains(l.IP)
}
