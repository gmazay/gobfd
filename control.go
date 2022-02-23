package gobfd

import (
	"fmt"
	"strings"
)

// Callback function
type CallbackFunc func(ipAddr string, preState, curState int) error

type Control struct {
	Local    string
	Family   int // ipv4, ipv6
	RxQueue  chan *RxData
	sessions []*Session
}

var log Logger

func NewControl(local string, family int, port int, logger Logger) *Control {
	controlPort = port
	log = logger
	tmpControl := &Control{
		Local:  local,
		Family: family,
		//log:    logger,
		RxQueue: make(chan *RxData),
	}
	tmpControl.Run()
	return tmpControl
}

////// 添加需要检测的实例 ///////
/*
 * local: local ip (0.0.0.0)
 * remote: peer ip
 * family: AF_INET4, AF_INET6
 * passive: Whether it is passive mode
 * rxInterval: receive interval (input in milliseconds),
 * txInterval: transmission interval (enter milliseconds)
 * detectMult: the maximum number of failed packets
 * f: Callback function
 */
func (c *Control) AddSession(remote string, passive bool, rxInterval, txInterval, detectMult int, f CallbackFunc) {
	nsession := NewSession(
		c.Local,
		remote,
		c.Family,
		passive,
		rxInterval*1000,
		txInterval*1000,
		detectMult,
		f,
	)
	//slogger.Debugf("Creating BFD session for remote %s.", remote)
	log.Printf("Creating BFD session for remote " + remote)
	c.sessions = append(c.sessions, nsession)
}

////// Delete an instance that needs to be detected  /////
func (c *Control) DelSession(remote string) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Del session error: %s", err)
			return
		}
	}()

	for i, session := range c.sessions {
		if session.Remote == remote {
			session.clientQuit <- true                               // Exit
			c.sessions = append(c.sessions[:i], c.sessions[i+1:]...) // Delete session
		}
	}

	return nil
}

// Process the received packet
func (c *Control) processPackets(rxdt *RxData) {
	//log.Printf("Received a new packet from %s.", rxdt.Addr)

	bfdPack := rxdt.Data
	if bfdPack.YourDiscriminator > 0 {
		for _, session := range c.sessions {
			if session.LocalDiscr == bfdPack.YourDiscriminator {
				session.RxPacket(bfdPack)
				return
			}
		}
	} else {
		for _, session := range c.sessions {
			//log.Printf("session remote: %v", rxdt.Addr)
			addrIp := strings.Split(rxdt.Addr, ":")[0]
			if session.Remote == addrIp {
				session.RxPacket(bfdPack)
				return
			}
		}
	}

	log.Printf("Dropping packet from %s as it doesnt match any configured remote.", rxdt.Addr)
}

func (c *Control) initServer() {
	log.Printf("Setting up udp server on %s:%d", c.Local, controlPort)
	addr := fmt.Sprintf("%s:%d", c.Local, controlPort)
	s := NewServer(addr, c.Family, c.RxQueue)
	go s.Start()

}

func (c *Control) backgroundRun() {
	c.initServer()
	log.Printf("BFD Daemon fully configured.")
	for {
		select {
		case rxData := <-c.RxQueue:
			c.processPackets(rxData)
		}
	}
}

func (c *Control) Run() {
	log.Printf("run...")
	go c.backgroundRun()
}
