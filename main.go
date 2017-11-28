package main

import (
	"log"
	"./router"
	"./reliable"
	"./client"
	"fmt"
	"net"
	"strconv"
	"strings"
	"flag"
	"./http"
	"net/url"
	"os"
)

func getPathDetails() (*url.URL, error){
	s := "http://127.0.0.1:8080/"
	if len(os.Args) >= 2 {
		s = os.Args[1];
	}

	return url.Parse(s)
}

func main() {
	go startServer()
	//dataPtr := flag.String("data", "default data", "data to send if using post")
	//headerPtr := flag.String("header", "", "headers to send")

	portReceivePtr := flag.Int("rPort", 8081, "Port to receive on")

	isGetPtr := flag.Bool("get", true, "do get")
	isPostPtr := flag.Bool("post", false, "do post")

	flag.Parse()

	u, _ := getPathDetails();
	fmt.Println(u)
	host, portStr, _ := net.SplitHostPort(u.Host);

	port, _ := strconv.Atoi(portStr);

	hostParts := strings.Split(host, ".")

	hostBytes := []byte{127, 0, 0, 1}
	fmt.Printf("%v \n", hostParts)
	if len(hostParts) > 0 {
		for i, part := range hostParts {
			asByte, _ := strconv.Atoi(part)
			hostBytes[i] = byte(asByte)
		}
	}

	//data := *dataPtr
	//headers := *headerPtr
	portReceive := *portReceivePtr
	isGet := *isGetPtr
	isPost := *isPostPtr

	requestType := http.GET

	if(isGet){
		requestType = http.GET
	}

	if(isPost){
		requestType = http.POST
	}


		client.Run(hostBytes, uint16(port), portReceive, requestType, u.Path)


}

func startServer(){
	routerConnection, err := router.ConnectToRouter("127.0.0.1", 3000,[]byte{127, 0, 0, 1}, 8081, 8080)

	if err != nil {
		log.Fatal(err)
	}
	defer routerConnection.Close()

		forwarder := reliable.CreateWindowForwarder(routerConnection)

		for true{
			windowConnection := forwarder.ReceiveAndForwardPacket()
			if windowConnection == nil{
				continue
			}
			if windowConnection.DoneSending {
				if len(windowConnection.SenderWindow.Buffer) > 0{
					if windowConnection.SenderWindow.Buffer[0].Packet == nil {
						windowConnection.Reset()
						continue
					}
				}else{
					windowConnection.Reset()
					continue
				}
			}

			if len(windowConnection.ReceiverWindow.InputBuffer) > 0 {
				data := windowConnection.GetWaitAndFlushInputBuffer()

				lastChar := len(data)-1
				for i := lastChar; i >= 0; i--{
					if(data[i] != 0){
						lastChar = i;
						break;
					}
				}

				text := string(data[:lastChar+1])

				windowConnection.SendPacket([]byte("Got the data on server: "+string(text)))
				windowConnection.DoneSending = true;
			}
		}

		fmt.Println("SERVER CLOSED")


}