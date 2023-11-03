package connection

import (
	"math/rand"
	"sync"
	"time"
)

type AddrBook interface {
	GetPeers() []Peer
	GetPeerByID(id string) (Peer, bool)
	AddPeer(Peer Peer) error
	RemovePeerByID(id string)
}

type DefaultAddrBook struct {
	mtx   sync.Mutex
	peers map[string]Peer
}

func NewAddrBook() AddrBook {
	return &DefaultAddrBook{
		peers: make(map[string]Peer),
	}
}

func (ab *DefaultAddrBook) GetPeers() []Peer {
	var peers []Peer
	for _, peer := range ab.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (ab *DefaultAddrBook) GetPeerByID(id string) (Peer, bool) {
	value, ok := ab.peers[id]
	return value, ok
}

func (ab *DefaultAddrBook) AddPeer(peer Peer) error {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(3)%int(3*time.Second/1000)))
	ab.mtx.Lock()
	defer ab.mtx.Unlock()
	if _, ok := ab.GetPeerByID(peer.(*DefaultPeer).ID); !ok {
		err := peer.Start()
		if err != nil {
			return err
		}
		ab.peers[peer.(*DefaultPeer).ID] = peer
	} else {
		err := peer.CloseStream()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ab *DefaultAddrBook) RemovePeerByID(id string) {
	delete(ab.peers, id)
}
