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

func (c *Controller) Run() (chan<- bool, *sync.WaitGroup) {
	stop := make(chan bool)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Println("Controller started")
		defer wg.Done()

		opts := mqtt.NewClientOptions()
		opts.AddBroker("tcp://dev:1883")

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Println("Error while connecting to broker: ", token.Error())
		}

		if token := client.Subscribe("/readings/temperature", 0, handlerReading); token.Wait() && token.Error() != nil {
			log.Println("Error while subscribing to topic: ", token.Error())
		}

		for {
			select {
			case <-stop:
				log.Println("Stopping controller")

				if token := client.Unsubscribe("/readings/temperature"); token.Wait() && token.Error() != nil {
					log.Println("Error unsubscribing from topic: ", token.Error())
				}

				client.Disconnect(1000)

				return
			default:
			}
		}
	}()

	return stop, &wg
}

// NewController creates a new controller and returns a pointer to it.
func NewController() *Controller {
	return &Controller{}
}
