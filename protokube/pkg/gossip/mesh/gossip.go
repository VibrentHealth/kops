package mesh

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/weaveworks/mesh"
	"k8s.io/kops/protokube/pkg/gossip"
	"net"
	"strconv"
	"time"
)

type MeshGossiper struct {
	seeds gossip.SeedProvider

	router *mesh.Router
	peer   *peer

	version uint64

	lastSnapshot *gossip.GossipStateSnapshot
}

func NewMeshGossiper(listen string, channelName string, nodeName string, password []byte, seeds gossip.SeedProvider) (*MeshGossiper, error) {
	meshConfig := mesh.Config{
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           password,
		ConnLimit:          64,
		PeerDiscovery:      true,
		//TrustedSubnets:     []*net.IPNet{},
	}

	{
		host, portString, err := net.SplitHostPort(listen)
		if err != nil {
			return nil, fmt.Errorf("cannot parse -listen flag: %v", listen)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			return nil, fmt.Errorf("cannot parse -listen flag: %v", listen)
		}
		meshConfig.Host = host
		meshConfig.Port = port
	}

	meshName, err := mesh.PeerNameFromUserInput(nodeName)
	if err != nil {
		return nil, fmt.Errorf("error parsing peer name: %v", err)
	}

	nickname := nodeName
	logger := &glogLogger{}
	router := mesh.NewRouter(meshConfig, meshName, nickname, mesh.NullOverlay{}, logger)

	peer := newPeer(meshName)
	gossip := router.NewGossip(channelName, peer)
	peer.register(gossip)

	gossiper := &MeshGossiper{
		seeds:  seeds,
		router: router,
		peer:   peer,
	}
	return gossiper, nil
}

func (g *MeshGossiper) Run() error {
	//glog.Infof("mesh router starting (%s)", *meshListen)
	g.router.Start()

	defer func() {
		glog.Infof("mesh router stopping")
		g.router.Stop()
	}()

	g.runSeeding()

	return nil
}

func (g *MeshGossiper) runSeeding() {
	for {
		glog.V(2).Infof("Querying for seeds")

		seeds, err := g.seeds.GetSeeds()
		if err != nil {
			glog.Warningf("error getting seeds: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		glog.Infof("Got seeds: %s", seeds)
		// TODO: Include ourselves?  Exclude ourselves?

		removeOthers := false
		errors := g.router.ConnectionMaker.InitiateConnections(seeds, removeOthers)

		if len(errors) != 0 {
			for _, err := range errors {
				glog.Infof("error connecting to seeds: %v", err)
			}

			time.Sleep(1 * time.Minute)
			continue
		}

		glog.V(2).Infof("Seeding successful")

		// Reseed periodically, just in case of partitions
		// TODO: Make it so that only one node polls, or at least statistically get close
		time.Sleep(60 * time.Minute)
	}
}

func (g *MeshGossiper) Snapshot() *gossip.GossipStateSnapshot {
	return g.peer.snapshot()
}

func (g *MeshGossiper) UpdateValues(removeKeys []string, putEntries map[string]string) error {
	glog.V(2).Infof("UpdateValues: remove=%s, put=%s", removeKeys, putEntries)
	return g.peer.updateValues(removeKeys, putEntries)
}
