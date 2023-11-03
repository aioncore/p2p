package services

import "fmt"

type ServiceData interface {
	GetHash() string
	GetAddress() string
	GetType() string
	String() string
}

type ServiceHead struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	SHA256     string `json:"sha256"`
	RPCAddress string `json:"rpc_address"`
}

func (s *ServiceHead) GetHash() string {
	return s.SHA256
}

func (s *ServiceHead) GetAddress() string {
	return s.RPCAddress
}

func (s *ServiceHead) GetType() string {
	return s.Type
}

func (s *ServiceHead) String() string {
	return fmt.Sprintf("%s service %s", s.Type, s.Name)
}
