package reliable

import ("log"
"../router"
)

const (
	DATA = 0
	ACK = 1
	NACK = 2
)


type TwoWayWindowConnection struct{
	SenderWindow Window
	ReceiverWindow Window
	Router *router.RouterConnection

	AcceptingPackets bool
}

func (twoWayWindowConnection *TwoWayWindowConnection)startReceiverServer(){
	go func() {
		var err error
		twoWayWindowConnection.Router, err = router.ConnectToRouter("127.0.0.1", 3000, 8080)

		if err != nil {
			log.Fatal(err)
		}

		for twoWayWindowConnection.AcceptingPackets{
			receivedPacket := twoWayWindowConnection.Router.ReadPacket()
			switch receivedPacket.Type{
			case DATA:
				twoWayWindowConnection.ReceiverWindow.onPacketReceived(&receivedPacket)
			case R_ACK:
			case R_NACK:
			case S_ACK:
			case S_NACK:
			}
		}
	}()
}

func (TwoWayWindowConnection *TwoWayWindowConnection) close(){
	TwoWayWindowConnection.AcceptingPackets = false;
	TwoWayWindowConnection.Router.Close()
}

type Window struct{
	BufferSize int
	Buffer []WindowElement
	AcceptedQueue []*Packet
}

type WindowElement struct{
	SequenceNumber uint32
	Packet *Packet
	canBeDiscarded bool
}

func (window *Window) MarkElementAsDiscardable(index int){

}

func (window *Window) onPacketReceived(packet *Packet){
	for i := 0; i < window.BufferSize; i++{
		bufferElement := &window.Buffer[i]
		if packet.SequenceNumber == bufferElement.SequenceNumber {
			bufferElement.canBeDiscarded = true
			bufferElement.Packet = packet

			for j := i; j >= 0; j--{
				bufferElementToCheck := window.Buffer[j]
				if(!bufferElementToCheck.canBeDiscarded){
					window.sendNackForSequenceNumber(bufferElementToCheck.SequenceNumber)
				}
			}

			break
		}
	}
	window.attemptToSlide()
}

func (window *Window) sendNackForSequenceNumber(sequenceNumber uint32){
	packet := CreatePacket(NACK, )
}


func(window *Window) attemptToSlide(){
	for window.Buffer[0].canBeDiscarded{
		window.AcceptedQueue = append(window.AcceptedQueue, window.Buffer[0].Packet)
		window.Buffer = window.Buffer[1:]
	}
}

