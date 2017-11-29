package reliable

import (
	"../router"
	"fmt"
	"strconv"
	"sync"
	"math/rand"
	"time"
)

type WindowForwarder struct{
	connections map[string]*TwoWayWindowConnection
	RouterConnection *router.RouterConnection
}

func CreateWindowForwarder(connection *router.RouterConnection) WindowForwarder{
	return WindowForwarder{make(map[string]*TwoWayWindowConnection), connection}
}

func CreateSenderWindow(name string, bufferSize uint32, routerConnection *router.RouterConnection) *SenderWindow{
	wnd :=  SenderWindow{name, bufferSize, make([]WindowSlot, bufferSize), make([]*router.Packet, 0), routerConnection, sync.Mutex{}, false}
	initialSequence := uint32(0)
	for i:= 0; i < int(wnd.BufferSize); i++{
		wnd.Buffer[i].SequenceNumber = initialSequence
		initialSequence++
	}
	return &wnd
}

func CreateReceiverWindow(name string, bufferSize uint32, routerConnection *router.RouterConnection) *ReceiverWindow{
	wnd :=  ReceiverWindow{name,make([]byte, 0), sync.Mutex{}, bufferSize, make([]WindowSlot, bufferSize), make([]*router.Packet, 0), routerConnection, sync.Mutex{}, false}
	initialSequence := uint32(0)
	for i:= 0; i < int(wnd.BufferSize); i++{
		wnd.Buffer[i].SequenceNumber = initialSequence
		initialSequence++
	}
	return &wnd
}

func (this *WindowForwarder) waitForSynchAck(connection *TwoWayWindowConnection, sequence uint32){
	syncPacket := connection.Router.CreatePacket(router.SYN, sequence, make([]byte, 0))
	connection.Router.SendPacket(&syncPacket)

	time.Sleep(time.Second*2)
	if !connection.ReceiverWindow.IsSynched {
		go connection.waitForSyncAck(sequence)
	}
}

func (windowForwarder *WindowForwarder) getConnectionForAddress(addr string, packet router.Packet) *TwoWayWindowConnection{
	connection, exist := windowForwarder.connections[addr]

	if !exist && packet.Type == router.SYN{
		windowSize := uint32(15)

		connection = &TwoWayWindowConnection{
 			CreateSenderWindow("Forward sender ", windowSize, windowForwarder.RouterConnection),
			CreateReceiverWindow("Forward receiver",windowSize, windowForwarder.RouterConnection),
			windowForwarder.RouterConnection, false, addr}

		connection.SenderWindow.IsSynched = true;


		ackPacket := connection.Router.CreatePacket(router.ACK, packet.SequenceNumber, make([]byte, 0))
		connection.Router.SendPacket(&ackPacket)

		sequenceStart := packet.SequenceNumber

		for i, _ := range connection.SenderWindow.Buffer{
			connection.SenderWindow.Buffer[i].SequenceNumber = sequenceStart+uint32(i);
		}

		fmt.Printf("Synched server sender window with sync id %v\n", connection.SenderWindow.Buffer[0].SequenceNumber)

		startSequence := rand.Uint32()%10


		for i, _ := range connection.ReceiverWindow.Buffer{
			connection.ReceiverWindow.Buffer[i].SequenceNumber = startSequence+uint32(i)
		}

		go windowForwarder.waitForSynchAck(connection, startSequence)

		fmt.Printf("Created new window connection for "+addr + " and synched at sender %v and receiver %v\n", sequenceStart, startSequence)

		windowForwarder.connections[addr] = connection
	}

	return connection
}

func (this *WindowForwarder) Reset(windowConnection *TwoWayWindowConnection){
	delete(this.connections, windowConnection.AddressStr)
}

func (windowForwarder *WindowForwarder) ReceiveAndForwardPacket() *TwoWayWindowConnection{

	receivedPacket := windowForwarder.RouterConnection.ReadPacket()


	ip := strconv.Itoa(int(receivedPacket.Address[0])) + strconv.Itoa(int(receivedPacket.Address[1])) + strconv.Itoa(int(receivedPacket.Address[2])) + strconv.Itoa(int(receivedPacket.Address[3]))
	lookupAddr := ip + ":" + strconv.Itoa(int(receivedPacket.Port))

	windowConnection := windowForwarder.getConnectionForAddress(lookupAddr, receivedPacket)

	if windowConnection == nil {
		return nil
	}

	if receivedPacket.Type == router.SYN {
		ackPacket := windowConnection.Router.CreatePacket(router.ACK, windowConnection.SenderWindow.Buffer[0].SequenceNumber, make([]byte, 0))
		windowConnection.Router.SendPacket(&ackPacket)
		return windowConnection
	}

	if !windowConnection.ReceiverWindow.IsSynched {
		if receivedPacket.Type == router.DATA {
			if receivedPacket.SequenceNumber == windowConnection.ReceiverWindow.Buffer[0].SequenceNumber {
				windowConnection.ReceiverWindow.IsSynched = true
				fmt.Printf("Synched server receiver window with sync id %v\n", windowConnection.ReceiverWindow.Buffer[0].SequenceNumber)
			}
		}
	}

	if windowConnection.ReceiverWindow.IsSynched {
		windowConnection.ProcessPacket(&receivedPacket)
	}

	return windowConnection;
}