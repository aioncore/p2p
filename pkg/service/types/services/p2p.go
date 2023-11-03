package services

type P2PServiceData struct {
	ServiceHead `json:"head"`
	Body        P2PServiceBody  `json:"body"`
	State       P2PServiceState `json:"state"`
}

type P2PServiceBody struct {
}

type P2PServiceState struct {
	Status string `json:"status"`
}

type P2PMessage []byte
