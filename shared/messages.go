package shared

import (
	"encoding/json"
	"time"
)

const (
	MSG_HB             = 0
	MSG_UPDATE_TARGETS = 1
)

const (
	AGENT_RESPONSE_FAILURE = 0
	AGENT_RESPONSE_SUCCESS = 1
)

type DcozMessage struct {
	MessageType   int
	SendResponse  bool
	ContainerIds  []string
	PausePeriod   time.Duration
	PauseDuration time.Duration
}

func MarshallDcozMessage(msg *DcozMessage) []byte {
	msgBytes, err := json.Marshal(*msg)
	if err != nil {
		// Message should be able to be marshalled
		panic(err)
	}
	return msgBytes
}

func UnmarshalDcozMessage(msgBytes []byte, output *DcozMessage) error {
	return json.Unmarshal(msgBytes, &output)
}
