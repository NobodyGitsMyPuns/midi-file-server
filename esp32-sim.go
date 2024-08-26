package main

import (
	"bufio"
	"fmt"
	"net"
)

// ConnectToESP32 connects to the ESP32 via TCP and sends a message
func ConnectToESP32(ip, message string) error {
	// Dial the TCP server on the ESP32's IP address and port 80
	conn, err := net.Dial("tcp", ip+":80")
	if err != nil {
		return fmt.Errorf("failed to connect to ESP32: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to ESP32 at %s\n", ip)

	// Send the message to the ESP32
	_, err = fmt.Fprintf(conn, message+"\n")
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	// Wait for the response from the ESP32
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("Received: %s", response)
	return nil
}
