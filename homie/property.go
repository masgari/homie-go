package homie

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Property homie node property
type Property interface {
	Name() string
	Type() string
	Value() string
	SetValue(value string) Property
	Node() Node
	SetNode(n Node) Property
	// Publish send current value as MQTT payload, topic will be Node().Topic(Name())
	Publish() Property

	// Subscribe called during initialisation, subscribe to MQTT topic: device/node/prop/set if property Handler is set
	Subscribe() Property

	Handler() PropertyHandler
	// SetHandler set handler for incomming MQTT messages, by setting Handler, the property will be settable (topic: device/node/prop/set)
	SetHandler(h PropertyHandler) Property
}

type property struct {
	name         string
	propertyType string
	value        string
	handler      PropertyHandler // if set, the property will be settable
	node         Node
}

func (p *property) Name() string {
	return p.name
}

func (p *property) Type() string {
	return p.propertyType
}

func (p *property) Value() string {
	return p.value
}

func (p *property) SetValue(value string) Property {
	p.value = value
	return p
}

func (p *property) Node() Node {
	return p.node
}

func (p *property) SetNode(n Node) Property {
	p.node = n
	return p
}
func (p *property) Handler() PropertyHandler {
	return p.handler
}
func (p *property) SetHandler(h PropertyHandler) Property {
	p.handler = h
	return p
}

func (p *property) Publish() Property {
	p.node.Device().SendMessage(p.Node().NodeTopic(p.name), p.value)
	return p
}

func (p *property) Subscribe() Property {
	if p.Handler() == nil {
		return p
	}
	topic := p.Node().Device().Topic(p.Node().NodeTopic(fmt.Sprintf("%s/set", p.name)))
	p.node.Device().Client().Subscribe(topic, 1, func(client mqtt.Client, message mqtt.Message) {
		p.onMessage(message.Topic(), message.Payload())
	})
	return p
}

func (p *property) onMessage(topic string, payload []byte) {
	if p.Handler() == nil {
		log.Fatalf("No handler for property: %s, topic: %s", p.name, topic)
		return
	}
	p.handler(p, payload, topic)
}
