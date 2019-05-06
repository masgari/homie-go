package homie

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// HomieSpecVersion Homie convention version
	HomieSpecVersion = "3.0.1"
)

// PropertyHandler a handler function type for a propery
type PropertyHandler func(p Property, payload []byte, topic string) (bool, error)

// MqttAdapter adapter for paho mqtt, to make it testable
type MqttAdapter interface {
	// IsConnected returns a bool signifying whether
	// the client is connected or not.
	IsConnected() bool

	// Publish will publish a message with the specified QoS and content
	// to the specified topic.
	// Returns a token to track delivery of the message to the broker
	Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token

	// Subscribe starts a new subscription. Provide a MessageHandler to be executed when
	// a message is published on the topic provided, or nil for the default handler
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token
}

type mqttClientDelegate struct {
	client mqtt.Client
}

func (a *mqttClientDelegate) IsConnected() bool {
	return a.client.IsConnected()
}

func (a *mqttClientDelegate) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return a.client.Publish(topic, qos, retained, payload)
}

func (a *mqttClientDelegate) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	return a.client.Subscribe(topic, qos, callback)
}
