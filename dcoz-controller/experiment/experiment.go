package experiment

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/NicholasSpringer/dcoz/dcoz-controller/tracker"
	"github.com/NicholasSpringer/dcoz/dcoz-controller/util"
	"github.com/NicholasSpringer/dcoz/shared"
)

const UPDATE_CONTAINER_IDS_PERIOD = time.Second * 10
const UPDATE_AGENTS_PERIOD = time.Second * 10
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
	exp.agents = make(map[string]*agentConn)
	agentIps, err := util.GetAgentIps()
	if err != nil {
		return nil, err
	}
	exp.connectToAgents(agentIps)
	if len(exp.agents) == 0 {
		return nil, errors.New("Experimenter cannot connect to any agents")
	}

	stats := []tracker.TrackerStats{}
	for _, target := range targets {
		// Check for new agents
		agentIps, err = util.GetAgentIps()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting agent IPs: %s\n", err)
		} else {
			exp.connectToAgents(agentIps)
		}
		// Update container ids
		err = exp.UpdateContainerIds(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting initial containerIds for target %s, skipping: %s\n", target, err)
			stats = append(stats, tracker.TrackerStats{})
			continue
		}

		hbTicker := time.NewTicker(shared.HEARTBEAT_PERIOD)
		updateAgentsTicker := time.NewTicker(UPDATE_AGENTS_PERIOD)
		updateContainerIdTicker := time.NewTicker(UPDATE_CONTAINER_IDS_PERIOD)
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
				agentIps, err = util.GetAgentIps()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting agent IPs: %s\n", err)
				} else {
					exp.connectToAgents(agentIps)
				}
			case <-updateContainerIdTicker.C:
				err = exp.UpdateContainerIds(target)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting containerIds for target %s: %s\n", target, err)
				}
			}
		}
		experimentStats := exp.tracker.FinishTracking()
		stats = append(stats, experimentStats)
	}
	return stats, nil
}

func (exp *Experimenter) UpdateContainerIds(target string) error {
	containerIds, err := util.GetContainerIds(target)
	if err != nil {
		return err
	}
	updateMsg := shared.DcozMessage{
		MessageType:   shared.MSG_UPDATE_EXP,
		ContainerIds:  containerIds,
		PausePeriod:   exp.config.PausePeriod,
		PauseDuration: exp.config.PauseDuration,
	}
	exp.broadcastMessage(&updateMsg)
	return nil
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
	agentIp string
	err     error
}

func (exp *Experimenter) broadcastMessage(msg *shared.DcozMessage) {
	msgBytes := shared.MarshallDcozMessage(msg)
	resultChan := make(chan agentMessageResult)
	// Multicast message
	for agentIp, agent := range exp.agents {
		go agent.sendMessage(msgBytes, agentIp, resultChan)
	}
	// Wait for messages to come back, delete agents if connection closed
	for i := 0; i < len(exp.agents); i++ {
		result := <-resultChan
		if result.err != nil {
			delete(exp.agents, result.agentIp)
			if result.err == net.ErrClosed {
				fmt.Fprintf(os.Stderr, "Connection closed to agent %s\n", result.agentIp)
			} else if result.err != nil {
				fmt.Fprintf(os.Stderr, "Connection error to agent %s: %s\n", result.agentIp, result.err)
			}
		}
	}
}

func (agent *agentConn) sendMessage(msg []byte, agentIp string, resultChan chan agentMessageResult) {
	_, err := agent.conn.Write(msg)
	result := agentMessageResult{
		agentIp: agentIp,
		err:     err,
	}
	resultChan <- result
}
