package controller

import (
	"encoding/json"
	"log"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Reading struct {
	SensorID    string  `json:"sensorID"`
	ReadingType string  `json:"type"`
	Value       float64 `json:"value"`
}

type Controller struct{}

var handlerReading mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	r := Reading{}
	err := json.Unmarshal(msg.Payload(), &r)
	if err != nil {
		log.Println("Error parsing JSON message: ", err)
		return
	}

	log.Printf("Received reading: sensor %v temp %v", r.SensorID, r.Value)
}

// Run starts the controller goroutine. It returns a quit channel, an error channel and a waitgroup
// for graceful shutdown.
func (c *Controller) Run() (chan<- bool, <-chan error, *sync.WaitGroup) {
	stop := make(chan bool)
	err := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Println("Controller started")
		defer wg.Done()

		// Set MQTT client options
		opts := mqtt.NewClientOptions()
		opts.AddBroker("tcp://dev:1883")

		// Connect to MQTT broker
		client := mqtt.NewClient(opts)
		log.Println("Connecting to MQTT broker")
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			err <- token.Error()
			return
		}
		log.Println("Connected to MQTT broker")

		defer func() {
			log.Println("Disconnecting from MQTT broker")
			client.Disconnect(1000)
			log.Println("Disconnected from MQTT broker")
		}()

		// Subscribe to readings topic
		if token := client.Subscribe("/readings/temperature", 0, handlerReading); token.Wait() && token.Error() != nil {
			err <- token.Error()
			return
		}

		// Wait for stop signal
		<-stop
		log.Println("Stopping controller")
		if token := client.Unsubscribe("/readings/temperature"); token.Wait() && token.Error() != nil {
			err <- token.Error()
		}
	}()

	return stop, err, &wg
}

// NewController creates a new controller and returns a pointer to it.
func NewController() *Controller {
	return &Controller{}
}
