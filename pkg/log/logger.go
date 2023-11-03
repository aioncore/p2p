package log

import (
	"github.com/aioncore/p2p/pkg/service/log"
)

func InitLogger(filePath string, serviceType string) {
	log.InitLogger(filePath, serviceType)
}
