package experiment

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/NicholasSpringer/dcoz/dcoz-controller/tracker"
	"github.com/NicholasSpringer/dcoz/dcoz-controller/util"
	"github.com/NicholasSpringer/dcoz/shared"
)

const UPDATE_AGENTS_PERIOD = time.Second * 5
const EXP_DURATION = time.Second * 60

type Experimenter struct {
	config  ExperimenterConfig
	tracker *tracker.Tracker
	agents  map[string]*agentConn
}

type ExperimenterConfig struct {
	PauseDuration time.Duration
	PausePeriod   time.Duration
}

func CreateExperimenter(config ExperimenterConfig, tracker *tracker.Tracker) *Experimenter {
	return &Experimenter{
		config:  config,
		tracker: tracker,
	}
}

func (exp *Experimenter) RunExperiments(targets []string) ([]tracker.TrackerStats, error) {
	fmt.Println("Experimenter started!")
	exp.agents = make(map[string]*agentConn)
	agentIps, err := util.GetAgentIps()
	if err != nil {
		return nil, err
	}
	exp.connectToAgents(agentIps)
	if len(exp.agents) == 0 {
		return nil, errors.New("Experimenter cannot connect to any agents")
	}
	fmt.Printf("Experimenter: Connected to %d agents\n", len(exp.agents))

	hbTicker := time.NewTicker(shared.HEARTBEAT_PERIOD)
	updateAgentsTicker := time.NewTicker(UPDATE_AGENTS_PERIOD)

	stats := []tracker.TrackerStats{}
	for _, target := range targets {
		// Check for new agents
		agentIps, err = util.GetAgentIps()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting agent IPs: %s\n", err)
		} else {
			exp.connectToAgents(agentIps)
		}
		fmt.Printf("Experimenter: synchronizing agents on target %s\n", target)
		// Update container ids, wait for response to synchronize
		err = exp.UpdateContainerIds(target, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting initial containerIds for target %s, skipping: %s\n", target, err)
			stats = append(stats, tracker.TrackerStats{})
			continue
		}

		endExperimentTimer := time.NewTimer(EXP_DURATION)
		exp.tracker.StartTracking()

	Exp:
		for {
			select {
			case <-endExperimentTimer.C:
				break Exp
			case <-hbTicker.C:
				msg := shared.DcozMessage{
					MessageType: shared.MSG_HB,
				}
				exp.broadcastMessage(&msg)
			case <-updateAgentsTicker.C:
				// Check for new agents
				fmt.Println("Experimenter: Getting agent IPs")
				agentIps, err = util.GetAgentIps()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting agent IPs: %s\n", err)
				} else {
					exp.connectToAgents(agentIps)
					exp.UpdateContainerIds(target, false)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error getting containerIds for target %s: %s\n", target, err)
					}
				}
			}
		}
		experimentStats := exp.tracker.FinishTracking()
		stats = append(stats, experimentStats)
	}
	exp.CloseAgentConns()
	return stats, nil
}

func (exp *Experimenter) UpdateContainerIds(target string, waitForResponse bool) error {
	containerIds, err := util.GetContainerIds(target)
	if err != nil {
		return err
	}
	updateMsg := shared.DcozMessage{
		MessageType:   shared.MSG_UPDATE_TARGETS,
		SendResponse:  waitForResponse,
		ContainerIds:  containerIds,
		PausePeriod:   exp.config.PausePeriod,
		PauseDuration: exp.config.PauseDuration,
	}
	if waitForResponse {
		fmt.Printf("Experimenter: synchronizing agents on new target %s\n", target)
	}
	exp.broadcastMessage(&updateMsg)
	return nil
}

func (exp *Experimenter) CloseAgentConns() {
	for agentIp, agent := range exp.agents {
		err := agent.close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error closing agent connection: %s\n", err)
		}
		delete(exp.agents, agentIp)
	}
}

type agentConn struct {
	conn net.Conn
}

func newAgentConn(agentIp string) (*agentConn, error) {
	agentAddr := fmt.Sprintf("%s:%d", agentIp, shared.AGENT_PORT)
	conn, err := net.Dial("tcp", agentAddr)
	if err != nil {
		return nil, err
	}
	agent := &agentConn{
		conn: conn,
	}
	return agent, nil
}

func (agent *agentConn) close() error {
	return agent.conn.Close()
}

type agentConnectResult struct {
	agent   *agentConn
	agentIp string
	err     error
}

func (exp *Experimenter) connectToAgents(agentIps []string) {
	newAgentIps := []string{}
	for _, agentIp := range agentIps {
		_, alreadyConnected := exp.agents[agentIp]
		if !alreadyConnected {
			newAgentIps = append(newAgentIps, agentIp)
		}
	}

	resultChan := make(chan agentConnectResult)
	// Multicast connection
	for _, agentIp := range newAgentIps {
		go connectToAgent(agentIp, resultChan)
	}
	// Wait for connections to be created, add to agents list
	for i := 0; i < len(newAgentIps); i++ {
		result := <-resultChan
		if result.err != nil {
			fmt.Fprintf(os.Stderr, "Could not connect to agent %s\n", result.agentIp)
		} else {
			exp.agents[result.agentIp] = result.agent
			fmt.Printf("Experimenter: Connected to agent %s\n", result.agentIp)
		}
	}
}

func connectToAgent(agentIp string, resultChan chan agentConnectResult) {
	agent, err := newAgentConn(agentIp)
	resultChan <- agentConnectResult{
		agent:   agent,
		agentIp: agentIp,
		err:     err,
	}
}

type agentMessageResult struct {
	agentIp  string
	err      error
	response byte
}

func (exp *Experimenter) broadcastMessage(msg *shared.DcozMessage) {
	msgBytes := shared.MarshallDcozMessage(msg)
	msgBytes = append(msgBytes, '\n')
	resultChan := make(chan agentMessageResult)
	// Multicast message
	for agentIp, agent := range exp.agents {
		go agent.sendMessage(msgBytes, agentIp, msg.SendResponse, resultChan)
	}
	// Wait for messages to come back, delete agents if connection closed
	for i := 0; i < len(exp.agents); i++ {
		result := <-resultChan
		if result.err != nil {
			delete(exp.agents, result.agentIp)
			if result.err == net.ErrClosed || result.err == io.EOF {
				fmt.Fprintf(os.Stderr, "Connection closed to agent %s\n", result.agentIp)
			} else if result.err != nil {
				fmt.Fprintf(os.Stderr, "Connection error to agent %s: %s\n", result.agentIp, result.err)
			}
		}
		if msg.SendResponse && result.response != shared.AGENT_RESPONSE_SUCCESS {
			fmt.Fprintf(os.Stderr, "Agent %s responded with failure\n", result.agentIp)
		}
	}
}

func (agent *agentConn) sendMessage(msg []byte, agentIp string, waitForResponse bool, resultChan chan agentMessageResult) {
	_, err := agent.conn.Write(msg)
	if err != nil {
		result := agentMessageResult{
			agentIp: agentIp,
			err:     err,
		}
		resultChan <- result
		return
	}
	if waitForResponse {
		resp := make([]byte, 1)
		n, err := agent.conn.Read(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response from agent %s: %s\n", agentIp, err)
			result := agentMessageResult{
				agentIp: agentIp,
				err:     err,
			}
			resultChan <- result
			return
		}
		if n != 1 {
			panic("Should not get more than one response")
		}
		result := agentMessageResult{
			agentIp:  agentIp,
			err:      nil,
			response: resp[0],
		}
		resultChan <- result
	} else {
		result := agentMessageResult{
			agentIp: agentIp,
			err:     nil,
		}
		resultChan <- result
	}
}
