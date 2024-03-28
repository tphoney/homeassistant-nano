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
	message string
	link    netlink.Netlinker
	sensor  = dht.New(pin, dht.DHT11)
	conn    net.Conn
	buf     = make([]byte, 256)
)

const (
	topic = "temperature/humidity"
	pin   = machine.D2
)

func main() {
	// Connect to WiFi
	connectWiFi()
	println("Connected to WiFi")
	time.Sleep(2 * time.Second)
	go func() {
		for {
			// Read temperature and humidity
			temp, hum, err := sensor.Measurements()
			if err != nil {
				println("Error reading DHT sensor:", err)
				continue
			}
			temp, hum = (temp / 10), (hum / 10)

			message = "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n" +
				`{"humidity":` + strconv.FormatInt(int64(hum), 10) + `, "temperature":` + strconv.FormatInt(int64(temp), 10) + "}\r\n"

			time.Sleep(time.Second * 30) // Adjust interval between readings
		}
	}()

	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		println("Error listening:", err)
		return
	}
	defer listener.Close()

	println("Server listening on port 80...")

	for {
		// Accept incoming connection
		conn, err = listener.Accept()
		if err != nil {
			println("Error accepting connection:", err)
			conn.Close()
			link.NetDisconnect()
			time.Sleep(5 * time.Second)
			connectWiFi()
			time.Sleep(5 * time.Second)
			continue
		}
		// Handle connection (simplified)
		err = handleHTTPConnection()
		if err != nil {
			link.NetDisconnect()
			time.Sleep(5 * time.Second)
			connectWiFi()
			time.Sleep(5 * time.Second)
		}
		conn.Close()
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

func handleHTTPConnection() (err error) {
	// Read request (simplified)
	_, err = conn.Read(buf)
	if err != nil {
		println("Error reading request:", err)
		return err
	}

	_, err = conn.Write([]byte(message))

	if err != nil {
		println("Error writing response:", err)
		return err
	}
	println("Request processed")
	return nil
}
