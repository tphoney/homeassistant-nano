package main

import (
	"context"
	"io"
	"machine"
	"net"
	"strconv"
	"time"

	mqtt "github.com/soypat/natiu-mqtt"
	"tinygo.org/x/drivers/dht"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid    string
	pass    string
	address string
	link    netlink.Netlinker
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
	time.Sleep(5 * time.Second)
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
	broker := address + ":1883"
	clientId := "arduino-tinygo-sensor"

	conn, err := net.Dial("tcp", broker)
	if err != nil {
		link.NetDisconnect()
		time.Sleep(5 * time.Second)
		connectWiFi()
		return err
	}
	defer conn.Close()

	println("Connected via TCP")
	client := mqtt.NewClient(mqtt.ClientConfig{
		Decoder: mqtt.DecoderNoAlloc{make([]byte, 1500)},
		OnPub: func(_ mqtt.Header, _ mqtt.VariablesPublish, r io.Reader) error {
			message, _ := io.ReadAll(r)
			println("received message:", string(message))
			return nil
		},
	})
	println("Client created")
	// Connect client
	var varConn mqtt.VariablesConnect
	println("1")
	varConn.SetDefaultMQTT([]byte(clientId))
	println("2")
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	println("3")
	err = client.Connect(ctx, conn, &varConn) // Connect to server.
	println("4")
	cancel()
	println("5")
	if err != nil {
		// Error or loop until connect success.
		println("connect attempt failed:", err)
		return err
	}
	println("Connected to MQTT broker")

	// // Publish on topic
	// pubFlags, _ := mqtt.NewPublishFlags(mqtt.QoS0, false, false)
	// pubVar := mqtt.VariablesPublish{
	// 	TopicName: []byte(topic),
	// }

	// err = client.PublishPayload(pubFlags, pubVar, []byte(message))
	// if err != nil {
	// 	println("failed to publish: ", err)
	// 	return err
	// }

	// time.Sleep(time.Second)

	// conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	client.Disconnect(err)
	return nil
}
