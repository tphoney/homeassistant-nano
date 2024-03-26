package main

import (
	"machine"
	"net"
	"strconv"
	"time"

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
	conn, err := net.Dial("tcp", address+":1883")
	if err != nil {
		link.NetDisconnect()
		time.Sleep(5 * time.Second)
		connectWiFi()
		return err
	}
	defer conn.Close()

	// Simple publish message (no QoS, retain)
	_, err = conn.Write([]byte("PUBLISH " + topic + "0 0\n" + message + "\n"))
	return err
}
