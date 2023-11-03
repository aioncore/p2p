package config

import (
	"fmt"
	"github.com/aioncore/p2p/pkg/service/config"
	"path/filepath"
	"reflect"
)

var P2P *P2PConfig

type P2PConfig struct {
	NodeKeyPath             string
	SelfMultiAddr           string
	BootstrapPeersMultiAddr []string
	ProtocolID              string
	Rendezvous              string
	RPCListenAddr           []string
}

func InitConfig(filePath string, serviceType string) {
	P2P = config.InitConfig(
		filePath,
		serviceType,
		DefaultCoreConfig(filePath),
		reflect.TypeOf(P2P),
	).(*P2PConfig)
}

func DefaultCoreConfig(filePath string) interface{} {
	return &P2PConfig{
		NodeKeyPath:             filepath.Join(filePath, "key", "node_key.json"),
		SelfMultiAddr:           fmt.Sprintf("/ip4/%s/tcp/%d", "0.0.0.0", 4001),
		BootstrapPeersMultiAddr: []string{},
		ProtocolID:              "/ouroboros_p2p/1.0.0",
		Rendezvous:              "ouroboros_p2p",
		RPCListenAddr:           []string{"tcp://127.0.0.1:26657"},
	}
}
