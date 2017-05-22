package baremetal

import (
	"k8s.io/kops/forks/pkg/dnsprovider"
	"k8s.io/kops/upup/pkg/fi"
)

type Cloud struct {
	dns dnsprovider.Interface
}

var _ fi.Cloud = &Cloud{}

func NewCloud(dns dnsprovider.Interface) (*Cloud, error) {
	return &Cloud{dns: dns}, nil
}

func (c *Cloud) ProviderID() fi.CloudProviderID {
	return fi.CloudProviderBareMetal
}

func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return c.dns, nil
}
