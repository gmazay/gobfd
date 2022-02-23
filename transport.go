package gobfd

import (
	"fmt"
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/google/gopacket/layers"
)

//const ( controlPort = 3784 )
var controlPort int

type RxData struct {
	Data *layers.BFD
	Addr string
}

type Client struct {
	Transport int
}

func RandInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func NewClient(local, remote string, family int) (*net.UDPConn, error) {
	var conn *net.UDPConn
	var err error
	var udpAddr *net.UDPAddr
	var rudpAddr *net.UDPAddr
	srcPort := RandInt(SourcePortMin, SourcePortMax)
	addr := fmt.Sprintf("%s:%d", local, srcPort)
	serAddr := fmt.Sprintf("%s:%d", remote, controlPort)
	if family == syscall.AF_INET6 {
		// ipv6
		udpAddr, _ = net.ResolveUDPAddr("udp6", addr)
		rudpAddr, _ = net.ResolveUDPAddr("udp6", serAddr)
	} else {
		// ipv4
		udpAddr, _ = net.ResolveUDPAddr("udp4", addr)
		rudpAddr, _ = net.ResolveUDPAddr("udp4", serAddr)
	}
	conn, err = net.DialUDP("udp", udpAddr, rudpAddr)
	if err != nil {
		return conn, err
	}
	return conn, nil
}

/////////////////////// server /////////////////////

type Server struct {
	Addr     string
	listener *net.UDPConn // udp conn

	Family  int
	RxQueue chan *RxData
}

func NewServer(addr string, family int, rx chan *RxData) *Server {
	return &Server{
		Addr:    addr,
		Family:  family,
		RxQueue: rx,
	}
}

func (s *Server) Start() error {
	if s.Family == syscall.AF_INET6 {
		// ipv6
		udpAddr, err := net.ResolveUDPAddr("udp6", s.Addr)
		if err != nil {
			log.Fatalf("ResolveUDPAddr err: " + err.Error())
			return err
		}
		s.listener, err = net.ListenUDP("udp6", udpAddr)
		if err != nil {
			log.Fatalf("ListenUDP err:" + err.Error())
			return err
		}
		log.Printf("udp server run at: %v", udpAddr.String())
	} else {
		// ipv4
		udpAddr, err := net.ResolveUDPAddr("udp4", s.Addr)
		if err != nil {
			log.Fatalf("ResolveUDPAddr err: " + err.Error())
			return err
		}
		s.listener, err = net.ListenUDP("udp4", udpAddr)
		if err != nil {
			log.Fatalf("ListenUDP err:" + err.Error())
			return err
		}
		log.Printf("udp server run at: %v", udpAddr.String())
	}

	defer s.listener.Close()

	s.Loop()

	return nil
}

func (s *Server) Loop() {
	for {
		data := make([]byte, 1024)
		n, udpConn, err := s.listener.ReadFromUDP(data)
		if err != nil {
			log.Printf("read from udp error:" + err.Error())
			continue
		}

		bfdPk, err := DecodePacket(data[:n])
		if err != nil {
			continue
		}
		rxData := &RxData{Data: bfdPk, Addr: udpConn.String()}
		s.RxQueue <- rxData
	}
}
