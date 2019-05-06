package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	homie "github.com/masgari/homie-go/homie"
)

var (
	publisher homie.PeriodicPublisher
)

func randomPropertySetter(n homie.Node) {
	valueProp := n.GetProperty("value")
	valueProp.SetValue(fmt.Sprintf("%d", rand.Intn(1000)))
	valueProp.Publish()
}

func periodicRandomIntPublisher(tickerPriod string) (homie.PeriodicPublisher, error) {
	interval, err := time.ParseDuration(tickerPriod)
	if err != nil {
		return nil, err
	}
	return homie.NewPeriodicPublisher(interval), nil
}

func main() {
	device := homie.NewDevice("test1", &homie.Config{
		Mqtt: homie.MqttConfig{
			Host:     "localhost",
			Port:     1883,
			Username: "user",
			Password: "password",
		},
		BaseTopic:           "devices/",
		StatsReportInterval: 60,
	})

	homie.NewDevicePublisher(device)

	publisher, _ = periodicRandomIntPublisher("1s")

	node := device.NewNode("RandomGenerator", "RandomValueGeneratorNode")

	publisher.AddNodePublisher(node, randomPropertySetter)

	node.NewProperty("value", "integer")

	// to change interval, send a message to: devices/test1/RandomGenerator/interval/set
	// sample intervals: 200ms, 3s
	node.NewProperty("interval", "integer").
		SetHandler(func(p homie.Property, payload []byte, topic string) (bool, error) {
			interval := string(payload)
			publisher.Close() // close current publisher
			newPublisher, err := periodicRandomIntPublisher(interval)
			if err != nil {
				log.Fatalf("Invalid ticker duration: %s, %v", interval, err)
				return false, err
			}
			publisher = newPublisher
			publisher.AddNodePublisher(node, randomPropertySetter)
			// invoke publisher again
			publisher.Start()
			return true, nil
		})
	device.Run(true)
}
