package homie

import (
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mqttTokenMock struct {
	mock.Mock
}

func (m *mqttTokenMock) Wait() bool {
	args := m.Called()
	return args.Get(0).(bool)
}
func (m *mqttTokenMock) WaitTimeout(time.Duration) bool {
	args := m.Called()
	return args.Get(0).(bool)
}
func (m *mqttTokenMock) Error() error {
	args := m.Called()
	return args.Get(0).(error)
}

type mqttAdapterMock struct {
	mock.Mock
}

func (m *mqttAdapterMock) IsConnected() bool {
	args := m.Called()
	return args.Get(0).(bool)
}
func (m *mqttAdapterMock) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	args := m.Called()
	return args.Get(0).(mqtt.Token)
}
func (m *mqttAdapterMock) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	args := m.Called(topic, qos, callback)
	//args[2].(mqtt.MessageHandler)()
	return args.Get(0).(mqtt.Token)
}

func makeTestDevice(name string) Device {
	return NewDevice(name, &Config{
		Mqtt: MqttConfig{
			Host:     "localhost",
			Port:     1883,
			Username: "user",
			Password: "password",
		},
		BaseTopic:           "devices/",
		StatsReportInterval: 60,
	})
}
func TestNewDevice(t *testing.T) {
	d := makeTestDevice("test1")
	assert.NotEqual(t, nil, d)
	assert.NotEqual(t, nil, d.Config())

	n1 := d.NewNode("n1", "Generic")
	d.NewNode("n2", "Generic")
	assert.NotEqual(t, nil, d.GetNode("n1"))
	assert.NotEqual(t, nil, d.GetNode("n1").Device())
	assert.NotEqual(t, nil, d.GetNode("n2"))
	assert.NotEqual(t, nil, d.GetNode("n2").Device())

	n1.NewProperty("p1", "integer")
	assert.NotEqual(t, nil, n1.GetProperty("p1"))
	assert.NotEqual(t, nil, n1.GetProperty("p1").Node())
	assert.NotEqual(t, nil, n1.GetProperty("p1").Node().Device())
}

func TestNodeTopic(t *testing.T) {
	d := makeTestDevice("test2")
	n := d.NewNode("n1", "Generic")
	assert.Equal(t, "n1/$name", n.NodeTopic("$name"))
}

func TestPropertyHandler(t *testing.T) {
	d := makeTestDevice("device-1")
	n1 := node{
		name: "n1",
	}
	d.AddNode(&n1)

	var (
		receivedPayload []byte
		topic           string
		prop            Property
	)

	handler := func(p Property, payload []byte, t string) (bool, error) {
		receivedPayload = payload
		topic = t
		prop = p
		p.SetValue(string(payload))
		return true, nil
	}
	p1 := &property{
		name: "p1",
	}
	n1.AddProperty(p1).
		SetHandler(handler)

	token := new(mqttTokenMock)
	client := new(mqttAdapterMock)
	client.On("IsConnected").Return(true).Once()
	// TODO: verify individual Publish calls by fixing m.Called() in mocked Publish() method and setup correct expectations
	client.On("Publish").Return(token).Times(8 + 3 + 1) // 8 device messages (1 publish stats) + 3 node messages + 1 propery value
	client.On("Subscribe", "devices/device-1/n1/p1/set", uint8(1), mock.AnythingOfType("mqtt.MessageHandler")).
		Return(token).
		Once()
	d.OnConnect(client)

	client.AssertExpectations(t)

	p1.onMessage("devices/device-1/n1/p1/set", []byte("new-value"))

	assert.Equal(t, []byte("new-value"), receivedPayload)
	assert.Equal(t, "devices/device-1/n1/p1/set", topic)
	assert.Equal(t, p1, prop)
	assert.Equal(t, "new-value", p1.Value())
}

func TestPeriodicPublisher(t *testing.T) {
	d := makeTestDevice("test-periodic-publisher")
	n := d.NewNode("n1", "Generic")

	var c1, c2 int
	p1 := NewPeriodicPublisher(time.Duration(8 * time.Millisecond))
	p1.AddNodePublisher(n, func(n Node) {
		t.Logf("c1: %d\n", c1)
		c1++
	})

	token := new(mqttTokenMock)
	client := new(mqttAdapterMock)
	client.On("IsConnected").Return(true)
	client.On("Publish").Return(token)
	client.On("Subscribe", mock.AnythingOfType("string"), uint8(1), mock.AnythingOfType("mqtt.MessageHandler")).
		Return(token)
	d.OnConnect(client)

	time.Sleep(100 * time.Millisecond)
	assert.True(t, c1 >= 9)

	// change period
	p2 := NewPeriodicPublisher(time.Duration(8 * time.Millisecond))
	defer p2.Close()
	p2.AddNodePublisher(n, func(n Node) {
		t.Logf("c2: %d\n", c2)
		c2++
	})
	p1.Close()

	n.NodePublisher()(n) // can use p2.Start()

	time.Sleep(100 * time.Millisecond)
	assert.True(t, c2 >= 9)
}
