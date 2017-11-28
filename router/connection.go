package router

import (
	"net"
	"encoding/binary"
	"strconv"
	"log"
//	"fmt"
)

type RouterInfo struct{
	Host string
	Port uint16

	endHost []byte
	endPort uint16

	Str string

}

type RouterConnection struct{
	info RouterInfo
	rawConnection net.PacketConn
}

func (routerConnection *RouterConnection) Close(){
	routerConnection.rawConnection.Close()
}

func (RouterConnection *RouterConnection) ReadPacket() (Packet){
	buffer := make([]byte, 1024)

	RouterConnection.rawConnection.ReadFrom(buffer)

	packetType := buffer[0]

	SequenceNum := binary.BigEndian.Uint32(buffer[1:5])

	data := buffer[11:]
	packet := RouterConnection.CreatePacket(packetType, SequenceNum, data)
	//fmt.Println("Read packet #", packet.SequenceNumber, " of type ", packet.Type, " from port ", packet.Port)

	return packet


}

func (routerConnection *RouterConnection) SendPacket(packet *Packet){

	//fmt.Println("Sending packet #", packet.SequenceNumber, " of type ", packet.Type, " to port ", packet.Port)

	var buf []byte

	buf = append(buf, packet.Type)

	sequenceBuffer := make([]byte, 4);
	binary.BigEndian.PutUint32(sequenceBuffer, packet.SequenceNumber)
	buf = append(buf, sequenceBuffer...)


	buf = append(buf, packet.Address...)

	portBuffer := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuffer, packet.Port)
	buf = append(buf, portBuffer...)

	buf = append(buf, packet.Data...)

	addr, err := net.ResolveUDPAddr("udp", routerConnection.info.Str)
	if err != nil {
		log.Fatal(err)
	}
	routerConnection.rawConnection.WriteTo(buf, addr)

}

func ConnectToRouter(routerHost string, routerPort uint16, host []byte, port uint16, receivePort int) (*RouterConnection, error){
	return connectToRouter(RouterInfo{routerHost, routerPort,host, port, routerHost+":"+strconv.Itoa(int(routerPort))}, receivePort)
}

func connectToRouter(info RouterInfo, receivePort int) (*RouterConnection, error){
	connection, error := net.ListenPacket("udp", "127.0.0.1:"+strconv.Itoa(receivePort))

	if error != nil{
		return nil, error;
	}

	return &RouterConnection{info, connection}, nil
}

type RawPacket struct{
	Host []byte
	Port uint16
}

func (router *RouterConnection) CreatePacket(Type uint8, SequenceNumber uint32, Data []byte) Packet {
	if router.info.endPort == 8081 {
		//fmt.Println("Kindw wortkd!?")
	}
	return Packet{Type, SequenceNumber, router.info.endHost, router.info.endPort, Data, router}

}