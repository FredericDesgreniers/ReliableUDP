package router

import (
	"net"
	"../reliable"
	"encoding/binary"
	"strconv"
)

type RouterInfo struct{
	Host string
	Port uint16

	Str string
}

type RouterConnection struct{
	info RouterInfo
	rawConnection net.PacketConn
}

func (routerConnection *RouterConnection) Close(){
	routerConnection.rawConnection.Close()
}

func (RouterConnection *RouterConnection) ReadPacket() (reliable.Packet){
	buffer := make([]byte, 1024)

	RouterConnection.rawConnection.ReadFrom(buffer)

	packetType := buffer[0]

	SequenceNum := binary.BigEndian.Uint32(buffer[1:5])
	PeerAddress := buffer[5:9]
	PortNumber := binary.BigEndian.Uint16(buffer[9:11])
	data := buffer[11:]

	return reliable.CreatePacket(packetType, SequenceNum, PeerAddress, PortNumber, data)


}

func (routerConnection *RouterConnection) SendPacket(packet *reliable.Packet){
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
	addr, _:= net.ResolveUDPAddr("udp", routerConnection.info.Str)

	routerConnection.rawConnection.WriteTo(buf, addr)

}

func ConnectToRouter(host string, port uint16, receivePort int) (*RouterConnection, error){
	return connectToRouter(RouterInfo{host, port, host+":"+strconv.Itoa(int(port))}, receivePort)
}

func connectToRouter(info RouterInfo, receivePort int) (*RouterConnection, error){
	connection, error := net.ListenPacket("udp", "127.0.0.1:"+strconv.Itoa(receivePort))

	if error != nil{
		return nil, error;
	}

	return &RouterConnection{info, connection}, nil
}

type Packet struct{
	Host []byte
	Port uint16
}