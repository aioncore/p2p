package p2p

import (
	"context"
	"fmt"
	"github.com/aioncore/p2p/pkg/config"
	"github.com/aioncore/p2p/pkg/p2p/connection"
	"github.com/aioncore/p2p/pkg/service"
	servicecrypto "github.com/aioncore/p2p/pkg/service/crypto"
	"github.com/aioncore/p2p/pkg/service/log"
	"github.com/aioncore/p2p/pkg/service/server/rpc/jsonrpc"
	"github.com/aioncore/p2p/pkg/service/types/services"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"net"
	"net/http"
	"strings"
	"sync"
)

type P2P struct {
	service.BaseService
	nodeKey         *servicecrypto.NodeKey
	serverListeners []net.Listener
	peerChan        chan peer.AddrInfo
	peerHost        host.Host
	kademliaDHT     *dht.IpfsDHT
	mdnsService     mdns.Service
	addrBook        connection.AddrBook
}

func NewP2P() *P2P {

	p2p := &P2P{
		peerChan: make(chan peer.AddrInfo),
		addrBook: connection.NewAddrBook(),
	}
	nodeKey, err := servicecrypto.LoadNodeKey(config.P2P.NodeKeyPath)
	if err != nil {
		panic(err)
	}
	p2p.nodeKey = nodeKey

	p2p.BaseService = *service.NewBaseService("p2p", p2p)

	prvKey, _, err := crypto.KeyPairFromStdKey(p2p.nodeKey.PrivateKey)
	if err != nil {
		panic(err)
	}
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(config.P2P.SelfMultiAddr)
	peerHost, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	p2p.peerHost = peerHost
	fmt.Println(peerHost.Addrs(), peerHost.ID())
	p2p.peerHost.SetStreamHandler(protocol.ID(config.P2P.ProtocolID), p2p.handleStream)
	ctx := context.Background()
	kademliaDHT, err := dht.New(ctx, p2p.peerHost)
	if err != nil {
		panic(err)
	}
	p2p.kademliaDHT = kademliaDHT
	p2p.mdnsService = mdns.NewMdnsService(p2p.peerHost, config.P2P.Rendezvous, p2p)

	return p2p
}

func (p2p *P2P) OnStart() error {
	err := p2p.bootstrapMDNS()
	if err != nil {
		return err
	}

	err = p2p.bootstrapDHT()
	if err != nil {
		return err
	}

	listeners, err := p2p.bootstrapRPCServer()
	if err != nil {
		return err
	}
	p2p.serverListeners = listeners
	return nil
}

func (p2p *P2P) bootstrapRPCServer() ([]net.Listener, error) {
	listeners := make([]net.Listener, len(config.P2P.RPCListenAddr))
	for i, listenAddr := range config.P2P.RPCListenAddr {
		mux := http.NewServeMux()
		p2pRoutes := GenerateRoutes(p2p)

		// JSONRPC endpoints
		mux.HandleFunc("/rpc", jsonrpc.MakeJSONRPCHandler(p2pRoutes))

		parts := strings.SplitN(listenAddr, "://", 2)
		proto, addr := parts[0], parts[1]
		listener, err := net.Listen(proto, addr)
		if err != nil {
			return nil, err
		}

		go func() {
			s := &http.Server{
				Handler: mux,
			}
			err := s.Serve(listener)
			if err != nil {
				return
			}
		}()

		listeners[i] = listener
	}
	return listeners, nil
}

func (p2p *P2P) OnStop() {
	for _, l := range p2p.serverListeners {
		log.Info("Closing rpc listener")
		if err := l.Close(); err != nil {
			log.Error("Error closing listener")
		}
	}
}

func (p2p *P2P) Name() string {
	return "p2p"
}

func (p2p *P2P) handleStream(stream network.Stream) {
	err := p2p.addrBook.AddPeer(connection.NewPeer(stream))
	if err != nil {
		log.Error(err.Error())
	}
}

// HandlePeerFound to be called when new  peer is found
func (p2p *P2P) HandlePeerFound(pi peer.AddrInfo) {
	p2p.peerChan <- pi
}

func (p2p *P2P) Broadcast(message services.P2PMessage) error {
	for _, p := range p2p.addrBook.GetPeers() {
		err := p.Send(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p2p *P2P) bootstrapMDNS() error {

	if err := p2p.mdnsService.Start(); err != nil {
		return err
	}
	go p2p.acceptPeerMDNS()
	return nil
}

func (p2p *P2P) acceptPeerMDNS() {
	ctx := context.Background()
	for { // allows multiple connection to join
		newPeer := <-p2p.peerChan // will block until we discover a peer
		log.Info("Found peer, connecting")

		if err := p2p.peerHost.Connect(ctx, newPeer); err != nil {
			log.Info("Connection failed")
			continue
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := p2p.peerHost.NewStream(ctx, newPeer.ID, protocol.ID(config.P2P.ProtocolID))
		if err != nil {
			log.Error("Stream open failed")
			continue
		}
		err = p2p.addrBook.AddPeer(connection.NewPeer(stream))
		if err != nil {
			continue
		}

	}
}

func (p2p *P2P) bootstrapDHT() error {
	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	log.Debug("Bootstrapping the DHT")
	ctx := context.Background()
	if err := p2p.kademliaDHT.Bootstrap(ctx); err != nil {
		return err
	}
	log.Info("begin to connect to bootstrap peers")
	var wg sync.WaitGroup
	for _, peerAddrString := range config.P2P.BootstrapPeersMultiAddr {
		peerAddr, err := multiaddr.NewMultiaddr(peerAddrString)
		if err != nil {
			return err
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			if err := p2p.peerHost.Connect(ctx, *peerInfo); err != nil {
				log.Info("Connect failed")
			} else {
				log.Info("Connection established with bootstrap node:", *peerInfo)
			}
		}()
	}
	wg.Wait()
	go p2p.acceptPeerDHT()
	return nil
}

func (p2p *P2P) acceptPeerDHT() {
	ctx := context.Background()
	log.Info("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(p2p.kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, config.P2P.Rendezvous)
	log.Debug("Successfully announced!")

	log.Debug("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, config.P2P.Rendezvous)
	if err != nil {
		panic(err)
	}

	for newPeer := range peerChan {
		if newPeer.ID == p2p.peerHost.ID() {
			continue
		}
		log.Debug("Found peer:", newPeer)

		if err := p2p.peerHost.Connect(ctx, newPeer); err != nil {
			log.Info("Connection failed")
			continue
		}

		// open a stream, this stream will be handled by handleStream other end
		stream, err := p2p.peerHost.NewStream(ctx, newPeer.ID, protocol.ID(config.P2P.ProtocolID))
		if err != nil {
			log.Error("Stream open failed")
			continue
		}
		err = p2p.addrBook.AddPeer(connection.NewPeer(stream))
		if err != nil {
			continue
		}
	}
}
