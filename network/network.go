package network

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/orochi-network/orochimaru/keypair"
	"github.com/orochi-network/orochimaru/logger"
	"go.uber.org/zap"
)

type Network struct {
	BindHost string
	BindPort uint
	NodeID   peer.ID
	Domain   string
	context  context.Context
	nodeKey  *keypair.KeyPair
	host     host.Host
	pubsub   *pubsub.PubSub
}

var log *zap.SugaredLogger

func init() {
	log = logger.GetSugarLogger()
}

func New(bindHost string, bindPort uint, domain string, nodeKey *keypair.KeyPair) *Network {
	bindStr := fmt.Sprintf("/ip4/%s/tcp/%d", bindHost, bindPort)
	log.Debugf("Bind address: %s", bindStr)
	sourceMultiAddr, err := multiaddr.NewMultiaddr(bindStr)
	if err != nil {
		log.Panic(err)
	}
	nodeID, _ := nodeKey.GetID()
	log.Debugf("Setup host with given private key, node ID: %s", nodeID)
	prvKey := nodeKey.GetPrivateKey()
	host, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		log.Panic(err)
	}

	context := context.Background()

	// Start new gossip pub sub
	pubsubInstance, err := pubsub.NewGossipSub(
		context,
		host,
		pubsub.WithPeerExchange(true),
	)

	if err != nil {
		log.Panic(err)
	}

	net := &Network{
		BindHost: bindHost,
		BindPort: bindPort,
		Domain:   domain,
		NodeID:   nodeID,
		nodeKey:  nodeKey,
		host:     host,
		context:  context,
		pubsub:   pubsubInstance,
	}

	return net
}

func (net *Network) Announce() {

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(net.context, net.host)
	if err != nil {
		log.Panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	log.Debug("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(net.context); err != nil {
		log.Panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := net.host.Connect(net.context, *peerinfo); err != nil {
				log.Warn(err)
			} else {
				log.Infof("Connection established with bootstrap node: %v", *peerinfo)
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point `domain` to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	log.Info("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(net.context, routingDiscovery, net.Domain)
	log.Debug("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	log.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(net.context, net.Domain)
	if err != nil {
		log.Panic(err)
	}

	for curPeer := range peerChan {
		if curPeer.ID == net.host.ID() {
			continue
		}

		log.Debugf("Connecting to: %s", curPeer.ID.Pretty())
		err := net.host.Connect(net.context, curPeer)

		if err != nil {
			log.Warnf("Connection failed: %v", err)
			continue
		}

		log.Infof("Connected to: %s", curPeer.ID.Pretty())
	}
}

func (net *Network) Join() {
	net.pubsub.GetTopics()

	topic, err := net.pubsub.Join("hello")

	if err != nil {
		log.Panic(err)
	}

	helloWorld, err := topic.Subscribe()

	if err != nil {
		log.Panic(err)
	}

	for {
		time.AfterFunc(time.Duration(rand.Intn(10))*time.Second, func() {
			topic.Publish(net.context, []byte("Hello from:"+net.NodeID.Pretty()))
		})
		msg, err := helloWorld.Next(net.context)
		if err != nil {
			log.Panic(err)
		}
		log.Debugf("Topic: %s from: %s data: %s", topic.String(), msg.GetFrom().String(), string(msg.GetData()))
	}
}
