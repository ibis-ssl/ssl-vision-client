package portmonitor

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslnet"
	"net"
	"sort"
	"sync"
	"time"
)

const activeThreshold = 2 * time.Second

type PortStatus struct {
	Port   int  `json:"port"`
	Active bool `json:"active"`
}

type ServiceStatus struct {
	Vision  []PortStatus `json:"vision"`
	Tracker []PortStatus `json:"tracker"`
	Referee []PortStatus `json:"referee"`
}

type portEntry struct {
	server       *sslnet.MulticastServer
	lastReceived time.Time
	mu           sync.Mutex
}

func (e *portEntry) touch() {
	e.mu.Lock()
	e.lastReceived = time.Now()
	e.mu.Unlock()
}

func (e *portEntry) isActive() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return time.Since(e.lastReceived) < activeThreshold
}

type Monitor struct {
	visionPorts  map[int]*portEntry
	trackerPorts map[int]*portEntry
	refereePorts map[int]*portEntry
}

func NewMonitor(visionPorts, trackerPorts, refereePorts []int, skipInterfaces []string) *Monitor {
	m := &Monitor{
		visionPorts:  make(map[int]*portEntry, len(visionPorts)),
		trackerPorts: make(map[int]*portEntry, len(trackerPorts)),
		refereePorts: make(map[int]*portEntry, len(refereePorts)),
	}

	const visionIP = "224.5.23.2"
	const trackerIP = "224.5.23.2"
	const refereeIP = "224.5.23.1"

	for _, p := range visionPorts {
		m.visionPorts[p] = newEntry(visionIP, p, skipInterfaces)
	}
	for _, p := range trackerPorts {
		m.trackerPorts[p] = newEntry(trackerIP, p, skipInterfaces)
	}
	for _, p := range refereePorts {
		m.refereePorts[p] = newEntry(refereeIP, p, skipInterfaces)
	}

	return m
}

func newEntry(ip string, port int, skipInterfaces []string) *portEntry {
	srv := sslnet.NewMulticastServer(fmt.Sprintf("%s:%d", ip, port))
	srv.SkipInterfaces = skipInterfaces
	e := &portEntry{server: srv}
	srv.Consumer = func(_ []byte, _ *net.UDPAddr) {
		e.touch()
	}
	return e
}

func (m *Monitor) Start() {
	for _, e := range m.visionPorts {
		e.server.Start()
	}
	for _, e := range m.trackerPorts {
		e.server.Start()
	}
	for _, e := range m.refereePorts {
		e.server.Start()
	}
}

func (m *Monitor) Stop() {
	for _, e := range m.visionPorts {
		e.server.Stop()
	}
	for _, e := range m.trackerPorts {
		e.server.Stop()
	}
	for _, e := range m.refereePorts {
		e.server.Stop()
	}
}

func (m *Monitor) Status() ServiceStatus {
	return ServiceStatus{
		Vision:  collectStatus(m.visionPorts),
		Tracker: collectStatus(m.trackerPorts),
		Referee: collectStatus(m.refereePorts),
	}
}

func collectStatus(entries map[int]*portEntry) []PortStatus {
	result := make([]PortStatus, 0, len(entries))
	for port, e := range entries {
		result = append(result, PortStatus{Port: port, Active: e.isActive()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Port < result[j].Port })
	return result
}
