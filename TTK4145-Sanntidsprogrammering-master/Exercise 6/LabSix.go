package main

import (
	// "os/exec"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type State struct {
	Number uint32
}

type Message struct {
	PrimaryState State
}

func Listen(incomingMessage chan Message) {
	local, err := net.ResolveUDPAddr("udp", ":33446")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", local)
	if err != nil {
		log.Fatal(err)
	}
	for {
		buffer := make([]byte, 1024)
		_, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println("Read", read_bytes, "from", sender)

		b := bytes.NewBuffer(buffer)
		msg := Message{}
		binary.Read(b, binary.BigEndian, &msg)
		incomingMessage <- msg
	}

}

func Send(outgoingMessage chan Message) {
	local, err := net.ResolveUDPAddr("udp", ":33000")
	if err != nil {
		log.Fatal(err)
	}

	bcast, err := net.ResolveUDPAddr("udp", "129.241.187.255:33446")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", local)
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg := <-outgoingMessage
		buffer := &bytes.Buffer{}
		binary.Write(buffer, binary.BigEndian, msg)

		_, err := conn.WriteToUDP(buffer.Bytes(), bcast)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println("Sent", sent_bytes)
		time.Sleep(1 * time.Second)
	}
}

func Master(initialState State, incomingMessage chan Message, outgoingMessage chan Message) {
	state := initialState
	for {
		time.Sleep(1 * time.Second)
		state.Number++
		time.Sleep(1 * time.Second)

		outgoingMessage <- Message{state}
		time.Sleep(1 * time.Second)

		fmt.Println("Master: ", state.Number)

		if state.Number == 5 {
			time.Sleep(8 * time.Second)
			break
		}

	}
}

func Backup(initialState State, incomingMessage chan Message, outgoingMessage chan Message) {
	state := initialState
Loop:
	for {
		select {
		case <-time.After(7 * time.Second):
			fmt.Println("Backup: ", state.Number)
			go Master(state, incomingMessage, outgoingMessage)
			break Loop
		case msg := <-incomingMessage:
			state = msg.PrimaryState
			fmt.Println("Backup up to date")
		}
	}
}

func main() {
	incomingMessage := make(chan Message)
	outgoingMessage := make(chan Message)

	go Listen(incomingMessage)
	go Send(outgoingMessage)

	initialState := State{0}

	go Master(initialState, incomingMessage, outgoingMessage)
	go Backup(initialState, incomingMessage, outgoingMessage)
	time.Sleep(50 * time.Second)
}
