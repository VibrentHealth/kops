package watchers

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/util"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

// NodeController watches for nodes
type NodeController struct {
	util.Stoppable
	kubeClient *client.CoreClient
	dns        dns.Context
}

// newNodeController creates a nodeController
func NewNodeController(kubeClient *client.CoreClient, dns dns.Context) (*NodeController, error) {
	c := &NodeController{
		kubeClient: kubeClient,
		dns:        dns,
	}

	c.dns.MarkReady("node", false)

	return c, nil
}

// Run starts the NodeController.
func (c *NodeController) Run() {
	glog.Infof("starting node controller")

	stopCh := c.StopChannel()
	go c.runWatcher(stopCh)

	<-stopCh
	glog.Infof("shutting down node controller")
}

func (c *NodeController) runWatcher(stopCh <-chan struct{}) {
	runOnce := func() (bool, error) {
		var listOpts api.ListOptions

		// Note we need to watch all the nodes, to set up alias targets
		listOpts.LabelSelector = labels.Everything()
		glog.Warningf("querying without field filter")
		listOpts.FieldSelector = fields.Everything()

		nodeList, err := c.kubeClient.Nodes().List(listOpts)
		if err != nil {
			return false, fmt.Errorf("error listing nodes: %v", err)
		}
		for i := range nodeList.Items {
			node := &nodeList.Items[i]
			glog.Infof("node: %v", node.Name)
			c.updateNodeRecords(node)
		}
		c.dns.MarkReady("node", true)

		// Note we need to watch all the nodes, to set up alias targets
		listOpts.LabelSelector = labels.Everything()
		glog.Warningf("querying without field filter")
		listOpts.FieldSelector = fields.Everything()

		listOpts.Watch = true
		listOpts.ResourceVersion = nodeList.ResourceVersion
		watcher, err := c.kubeClient.Nodes().Watch(listOpts)
		if err != nil {
			return false, fmt.Errorf("error watching nodes: %v", err)
		}
		ch := watcher.ResultChan()
		for {
			select {
			case <-stopCh:
				glog.Infof("Got stop signal")
				return true, nil
			case event, ok := <-ch:
				if !ok {
					glog.Infof("node watch channel closed")
					return false, nil
				}

				node := event.Object.(*v1.Node)
				glog.V(4).Infof("node changed: %s %v", event.Type, node.Name)

				switch event.Type {
				case watch.Added, watch.Modified:
					c.updateNodeRecords(node)

				case watch.Deleted:
					c.dns.Replace("node", node.Name, nil)
				}
			}
		}
	}

	for {
		stop, err := runOnce()
		if stop {
			return
		}

		if err != nil {
			glog.Warningf("Unexpected error in event watch, will retry: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
}

func (c *NodeController) updateNodeRecords(node *v1.Node) {
	var records []dns.Record

	//dnsLabel := node.Labels[LabelNameDns]
	//if dnsLabel != "" {
	//	var ips []string
	//	for _, a := range node.Status.Addresses {
	//		if a.Type != v1.NodeExternalIP {
	//			continue
	//		}
	//		ips = append(ips, a.Address)
	//	}
	//	tokens := strings.Split(dnsLabel, ",")
	//	for _, token := range tokens {
	//		token = strings.TrimSpace(token)
	//
	//		// Assume a FQDN A record
	//		fqdn := token
	//		for _, ip := range ips {
	//			records = append(records, dns.Record{
	//				RecordType: dns.RecordTypeA,
	//				FQDN: fqdn,
	//				Value: ip,
	//			})
	//		}
	//	}
	//}
	//
	//dnsLabelInternal := node.Annotations[AnnotationNameDnsInternal]
	//if dnsLabelInternal != "" {
	//	var ips []string
	//	for _, a := range node.Status.Addresses {
	//		if a.Type != v1.NodeInternalIP {
	//			continue
	//		}
	//		ips = append(ips, a.Address)
	//	}
	//	tokens := strings.Split(dnsLabelInternal, ",")
	//	for _, token := range tokens {
	//		token = strings.TrimSpace(token)
	//
	//		// Assume a FQDN A record
	//		fqdn := dns.EnsureDotSuffix(token)
	//		for _, ip := range ips {
	//			records = append(records, dns.Record{
	//				RecordType: dns.RecordTypeA,
	//				FQDN: fqdn,
	//				Value: ip,
	//			})
	//		}
	//	}
	//}

	// Alias targets

	// node/<name>/internal -> InternalIP
	for _, a := range node.Status.Addresses {
		if a.Type != v1.NodeInternalIP {
			continue
		}
		records = append(records, dns.Record{
			RecordType:  dns.RecordTypeA,
			FQDN:        "node/" + node.Name + "/internal",
			Value:       a.Address,
			AliasTarget: true,
		})
	}

	// node/<name>/external -> ExternalIP
	for _, a := range node.Status.Addresses {
		if a.Type != v1.NodeExternalIP {
			continue
		}
		records = append(records, dns.Record{
			RecordType:  dns.RecordTypeA,
			FQDN:        "node/" + node.Name + "/external",
			Value:       a.Address,
			AliasTarget: true,
		})
	}

	c.dns.Replace("node", node.Name, records)
}
