package main

import (
	"machine"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"tinygo.org/x/drivers/dht"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid              string
	pass              string
	address           string
	link              netlink.Netlinker
	messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		println("Message " + string(msg.Payload()) + ". Topic " + msg.Topic())
	}

	connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		println("Connected")
	}

	connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		println("Connection Lost: ", err.Error())
	}
)

const (
	topic = "temperature/humidity"
)

func main() {
	// Configure DHT sensor pin (replace with your actual pin)
	pin := machine.D2
	sensor := dht.New(pin, dht.DHT11)
	// Connect to WiFi
	connectWiFi()
	println("Connected to WiFi")

	for {
		// Read temperature and humidity
		temp, hum, err := sensor.Measurements()
		if err != nil {
			println("Error reading DHT sensor:", err)
			continue
		}
		temp, hum = (temp / 10), (hum / 10)
		// Prepare MQTT message
		message := `{"humidity":` + strconv.FormatInt(int64(hum), 10) + `"temperature"` + strconv.FormatInt(int64(temp), 10) + `}`
		// Publish MQTT message
		err = publishMQTT(message)
		if err != nil {
			println("Error publishing MQTT:", err)
			continue
		}
		println("Sent MQTT message:", message)
		time.Sleep(time.Second * 10) // Adjust interval between readings
	}
}

func connectWiFi() {
	link, _ = probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		println(err.Error())
		time.Sleep(5 * time.Second)
		connectWiFi()
	}
}

func publishMQTT(message string) error {
	broker := "tcp://" + address + ":1883"
	clientId := "tinygo-client"
	options := mqtt.NewClientOptions()

	options.AddBroker(broker)
	options.SetClientID(clientId)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	println("Connecting to MQTT broker at ", broker)
	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		sleep := 5 * time.Second
		link.NetDisconnect()
		time.Sleep(sleep)
		connectWiFi()
		return token.Error()
	}

	token = client.Publish(topic, 0, false, message)
	time.Sleep(1 * time.Second)
	client.Disconnect(250)
	return token.Error()
}
