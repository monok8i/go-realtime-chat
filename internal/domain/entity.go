package domain

type Payload struct {
	UserID  int    `json:"user_id"`
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
}

type IncomingBrokerMessage struct {
	Body []byte
	Ack  func() error
}

type IncomingPubSubMessage struct {
	Payload []byte
}
