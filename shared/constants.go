package shared

import "time"

const (
	AGENT_PORT        = 8081
	AGENT_BUF_SIZE    = 4096
	HEARTBEAT_PERIOD  = time.Second * 1
	HEARTBEAT_TIMEOUT = HEARTBEAT_PERIOD * 3
)
