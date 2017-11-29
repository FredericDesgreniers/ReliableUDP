package main

import (
	"./client"
	"fmt"
	"net"
	"strconv"
	"strings"
	"flag"
	"./http"
	"net/url"

	"time"
	"math/rand"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	dataPtr := flag.String("data", "", "data to send if using post")
	//headerPtr := flag.String("header", "", "headers to send")

	portReceivePtr := flag.Int("rPort", 8081, "Port to receive on")

	isGetPtr := flag.Bool("get", true, "do get")
	isPostPtr := flag.Bool("post", false, "do post")

	pathPtr := flag.String("p", "http://127.0.0.1:8080/", "request path")

	flag.Parse()

	u, _ := url.Parse(*pathPtr)

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
	data := *dataPtr

	fmt.Println(isPost)

	requestType := http.GET

	if isGet {
		requestType = http.GET
	}

	if isPost {
		requestType = http.POST
	}


		client.Run(hostBytes, uint16(port), portReceive, requestType, u.Path, data)


}

