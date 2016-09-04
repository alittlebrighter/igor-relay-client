package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alittlebrighter/treehouse-relay-client"
)

func main() {
	id := flag.String("id", "pi-0", "Identification to be used on the server.")
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	key := flag.String("key", "AES256Key-32Characters1234567890", "The symmetric key to use.")
	flag.Parse()

	rClient := relayClient.NewRelayClient(*id, *host, []byte(*key))

	for {
		msg := new(Message)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Destination Controller: ")
		msg.Controller, _ = reader.ReadString('\n')
		msg.Controller = strings.TrimSpace(msg.Controller)
		fmt.Print("Module to control: ")
		msg.Instruction.Module, _ = reader.ReadString('\n')
		msg.Instruction.Module = strings.TrimSpace(msg.Instruction.Module)
		fmt.Print("Set to mode: ")
		msg.Instruction.Mode, _ = reader.ReadString('\n')
		msg.Instruction.Mode = strings.TrimSpace(msg.Instruction.Mode)
		_, err := rClient.SendMessage(msg.Controller, msg, json.Marshal)
		if err != nil {
			log.Printf("Error sending message: %s\n", err.Error())
		}
	}
}

type Message struct {
	Controller  string
	Instruction struct {
		Module string
		Mode   string
	}
}
