package connection

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/aioncore/p2p/pkg/service/log"
	"github.com/aioncore/p2p/pkg/service/types/services"
	"github.com/libp2p/go-libp2p/core/network"
	"time"
)

type Peer interface {
	Start() error
	sendStream(rw *bufio.ReadWriter)
	receiveStream(rw *bufio.ReadWriter)
	Send(message services.P2PMessage) error
	CloseStream() error
}

type DefaultPeer struct {
	ID              string
	stream          network.Stream
	pingTimer       *time.Ticker
	pongTimer       *time.Timer
	pongTimeoutChan chan bool
	pongChan        chan struct{}
	sendChan        chan interface{}
}

func NewPeer(stream network.Stream) Peer {
	return &DefaultPeer{
		stream: stream,
	}
}

func (p *DefaultPeer) Start() error {
	rw := bufio.NewReadWriter(bufio.NewReader(p.stream), bufio.NewWriter(p.stream))
	p.pingTimer = time.NewTicker(60 * time.Second)
	p.pongChan = make(chan struct{})
	p.pongTimeoutChan = make(chan bool, 1)
	go p.sendStream(rw)
	go p.receiveStream(rw)
	return nil
}

func (p *DefaultPeer) CloseStream() error {
	err := p.stream.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *DefaultPeer) sendStream(rw *bufio.ReadWriter) {
	for {
		select {
		case <-p.pingTimer.C:
			log.Debug("Send Ping")
			data := "ping"
			byteData := []byte(data)
			var length int32
			length = int32(len(byteData))

			err := binary.Write(rw, binary.LittleEndian, length)
			if err != nil {
				panic(err)
			}
			err = binary.Write(rw, binary.LittleEndian, byteData)
			if err != nil {
				panic(err)
			}
			p.pongTimer = time.AfterFunc(45*time.Second, func() {
				select {
				case p.pongTimeoutChan <- true:
				default:
				}
			})
			err = rw.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}
		case message := <-p.sendChan:
			_, err := rw.Write(message.([]byte))
			if err != nil {
				return
			}
			err = rw.Flush()
			if err != nil {
				fmt.Println("Error flushing buffer")
				panic(err)
			}
		}

	}
}

func (p *DefaultPeer) receiveStream(rw *bufio.ReadWriter) {
	for {
		var length int32
		err := binary.Read(rw, binary.LittleEndian, &length)
		if err != nil {
			panic(err)
		}
		data := make([]byte, length)
		err = binary.Read(rw, binary.LittleEndian, data)
		if err != nil {
			panic(err)
		}
		log.Info("received %s", string(data))
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

	}
}

func (p *DefaultPeer) Send(message services.P2PMessage) error {
	select {
	case p.sendChan <- message:
	default:
	}
	return nil
}
