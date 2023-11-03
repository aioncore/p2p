package p2p

import (
	"github.com/aioncore/p2p/pkg/service/server/utils"
)

func GenerateRoutes(p2p *P2P) map[string]*utils.APIFunc {
	routes := map[string]*utils.APIFunc{
		"broadcast": utils.NewAPIFunc(p2p.Broadcast, "message"),
	}
	return routes
}
