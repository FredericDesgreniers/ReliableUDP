package client

import (
	"../router"
	"../reliable"
	"fmt"
	"log"
	"../http"
)

func createRequestLine(method int, path string) string{
	line := "";

	switch(method){
	case http.GET:
		line += "GET";
		break
	case http.POST:
		line += "POST";
		break;
	}

	line += " " + path



	return line;
}



func Run(hostBytes []byte, port uint16, portReceive int, requestType int, path string) {
	routerConnection, err := router.ConnectToRouter("127.0.0.1", 3000, hostBytes, port, portReceive)

	if err != nil {
		log.Fatal(err)
	}
	defer routerConnection.Close()

	windowSize := uint32(15)

	connection := &reliable.TwoWayWindowConnection{
		reliable.CreateSenderWindow("Client sender", windowSize, routerConnection),
		reliable.CreateReceiverWindow("Client receiver", windowSize, routerConnection),
		routerConnection, false}

	connection.Connect()

	fmt.Println("Sending " +createRequestLine(requestType, path ))
	connection.SendPacket([]byte(createRequestLine(requestType, path)))

	for true {
		receivedPacket := routerConnection.ReadPacket()
		connection.ProcessPacket(&receivedPacket)

		if len(connection.ReceiverWindow.InputBuffer) > 0 {
			data := connection.GetWaitAndFlushInputBuffer()
			if len(data) > 0 {
				fmt.Println("Data received on client: " + string(data))
				routerConnection.Close()
				return;
			}
		}

	}

}

