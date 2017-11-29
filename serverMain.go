package main

import (
	"./router"
	"./reliable"
	"fmt"
	"strings"
	"io/ioutil"
	"log"
)

func main()  {
	startServer()
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
					forwarder.Reset(windowConnection)
					fmt.Println("Resetting connection!")
					continue
				}
			}else{

				forwarder.Reset(windowConnection)
				fmt.Println("Resetting connection!")
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

			fmt.Println(text)

			textLines := strings.Split(text, "\n")
			requestArgs := strings.Split(textLines[0], " ")

			requestType := requestArgs[0]
			requestPath := requestArgs[1]

			fmt.Printf("Got request for %v @ %v\n", requestType, requestPath)

			response := []byte("response... ")

			switch requestType{
			case "GET":
				if requestPath == "/" {
					dir , _:= ioutil.ReadDir("./")
					for _, f := range dir {
						if !f.IsDir() {
							response = append(response, []byte(f.Name() + "\n")...)
						}
					}
				}else {
					file, err := ioutil.ReadFile("."+requestPath)
					if err == nil {
						response = append(response, file...)
						fmt.Println("Sending file " + string(file))
					} else{
						response = append(response, []byte("File not here, refer to /")...)
					}
				}
				break
			case "POST":
				response = append(response, []byte("Writing data to file!")...)
				fileContents := []byte(strings.Join(textLines[1:], ""))
				ioutil.WriteFile("."+requestPath, fileContents, 0644)
				break
			default:
				response = append(response, []byte("Invalid request")...)
			}

			windowConnection.SendPacket([]byte(response))
			windowConnection.DoneSending = true;
		}
	}

	fmt.Println("SERVER CLOSED")


}