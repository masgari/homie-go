package homie

import (
	"sync"
	"time"
)

// NodePublisher publish node properties, it will be called during device initialisation
type NodePublisher func(n Node)

// DevicePublisher publish device stats
type DevicePublisher func(d Device)

// PeriodicPublisher periodically invoke configured publishers, can have multiple instances of PeriodicPublisher
// for example, group some nodes to publish properties every minutes and some other nodes to publish every hour
// device can have only one publisher, if multiple PeriodicPublisher configured for a device, there will be a panic
type PeriodicPublisher interface {
	GetDevicePublisher() DevicePublisher
	SetDevicePublisher(d Device, publisher DevicePublisher) PeriodicPublisher
	GetNodePublisher(node Node) NodePublisher
	AddNodePublisher(node Node, publisher NodePublisher) PeriodicPublisher
	Start()
	Close()
}

type periodicPublisher struct {
	devicePublisher DevicePublisher
	device          Device
	nodePublishers  map[Node]NodePublisher
	ticker          *time.Ticker
	done            chan bool
	started         bool
	mutex           *sync.Mutex
}

func (p *periodicPublisher) GetDevicePublisher() DevicePublisher {
	return p.devicePublisher
}

func (p *periodicPublisher) SetDevicePublisher(d Device, publisher DevicePublisher) PeriodicPublisher {
	p.devicePublisher = publisher
	p.device = d
	p.device.SetDevicePublisher(func(d Device) {
		p.Start()
	})
	return p
}

func (p *periodicPublisher) AddNodePublisher(node Node, publisher NodePublisher) PeriodicPublisher {
	p.nodePublishers[node] = publisher
	node.SetNodePublisher(func(n Node) {
		p.Start()
	})
	return p
}

func (p *periodicPublisher) GetNodePublisher(node Node) NodePublisher {
	return p.nodePublishers[node]
}

func (p *periodicPublisher) Start() {
	p.mutex.Lock()
	if p.started {
		defer p.mutex.Unlock()
		return
	}
	go func() {
		p.started = true
		p.mutex.Unlock()
		for {
			select {
			case <-p.done:
				return
			case <-p.ticker.C:
				p.invokePublishers()
			}
		}
	}()
}
func (p *periodicPublisher) invokePublishers() {
	if p.devicePublisher != nil {
		p.devicePublisher(p.device)
	}
	for node, nodePublisher := range p.nodePublishers {
		nodePublisher(node)
	}
}

func (p *periodicPublisher) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.ticker.Stop()
	// used a goroutine to avoid blocking in case of publisher routine is crashed or so
	go func() { p.done <- true }()
	p.started = false
}

// NewPeriodicPublisher create PeriodicPublisher
func NewPeriodicPublisher(period time.Duration) PeriodicPublisher {
	return &periodicPublisher{
		nodePublishers: make(map[Node]NodePublisher),
		done:           make(chan bool),
		ticker:         time.NewTicker(period),
		started:        false,
		mutex:          &sync.Mutex{},
	}
}

// NewDevicePublisher create default device publisher to publish device stats (uptime)
func NewDevicePublisher(d Device) PeriodicPublisher {
	p := NewPeriodicPublisher(time.Duration(d.Config().StatsReportInterval) * time.Second)
	p.SetDevicePublisher(d, func(d Device) {
		d.PublishStats()
	})
	return p
}
