package main

import (
	"log"
	"./router"
	"./reliable"
	"fmt"
)

func main() {
	go startServer()

	routerConnection, err := router.ConnectToRouter("127.0.0.1", 3000, 8081)

	if err != nil {
		log.Fatal(err)
	}
	defer routerConnection.Close()

	packet := reliable.CreatePacket(0, 0, []byte{127, 0, 0, 1}, 8080, []byte("This is data"))
	routerConnection.SendPacket(&packet)



	for 1 == 1{

	}
}

func startServer(){
	routerConnection, err := router.ConnectToRouter("127.0.0.1", 3000, 8080)

	if err != nil {
		log.Fatal(err)
	}
	defer routerConnection.Close()


	receivedPacket := routerConnection.ReadPacket()

	fmt.Printf("Received from %v:%v a packet with data \"%s\"", receivedPacket.Address, receivedPacket.Port, string(receivedPacket.Data))
}