package homie

import (
	"fmt"
	"log"
	"strings"
)

// Node homie node type
type Node interface {
	Name() string
	Type() string
	Device() Device
	SetDevice(d Device) Node
	NewProperty(name string, propertyType string) Property
	AddProperty(p Property) Property
	GetProperty(name string) Property
	NodePublisher() NodePublisher
	SetNodePublisher(publisher NodePublisher) Node

	// NodeTopic returns relative topic name for a part, for example timeNode/currentTime
	NodeTopic(part string) string

	Publish() Node
	// Subscribe subscribe node properties
	Subscribe() Node
}

type node struct {
	id         string
	name       string
	nodeType   string
	device     Device
	properties map[string]Property
	publisher  NodePublisher
}

func (n *node) Name() string {
	return n.name
}
func (n *node) Type() string {
	return n.nodeType
}
func (n *node) Device() Device {
	return n.device
}
func (n *node) SetDevice(d Device) Node {
	n.device = d
	return n
}
func (n *node) NodePublisher() NodePublisher {
	return n.publisher
}
func (n *node) SetNodePublisher(publisher NodePublisher) Node {
	n.publisher = publisher
	return n
}

func (n *node) GetProperty(name string) Property {
	return n.properties[name]
}

func (n *node) NewProperty(name string, propertyType string) Property {
	return n.AddProperty(&property{
		name:         name,
		propertyType: propertyType,
	})
}

func (n *node) AddProperty(p Property) Property {
	p.SetNode(n)
	if n.properties == nil {
		n.properties = make(map[string]Property)
	}
	if _, alreadyAdded := n.properties[p.Name()]; alreadyAdded {
		log.Panic(fmt.Errorf("Property %s already added to node: %s", p.Name(), n.name))
	}
	n.properties[p.Name()] = p
	return p
}

func (n *node) NodeTopic(part string) string {
	return fmt.Sprintf("%s/%s", n.name, part)
}

func (n *node) Subscribe() Node {
	for _, p := range n.properties {
		p.Subscribe()
	}
	return n
}

func (n *node) Publish() Node {
	n.device.SendMessage(n.NodeTopic("$name"), n.name)
	n.device.SendMessage(n.NodeTopic("$type"), n.nodeType)
	var propNames []string
	for _, p := range n.properties {
		propNames = append(propNames, p.Name())
	}
	n.Device().SendMessage(n.NodeTopic("$properties"), strings.Join(propNames, ","))
	for _, p := range n.properties {
		p.Publish()
	}
	return n
}
