package reliable

import (
	"../router"
	"runtime"

	"math/rand"
	"time"
	"fmt"
)

type TwoWayWindowConnection struct{
	SenderWindow *SenderWindow
	ReceiverWindow *ReceiverWindow
	Router *router.RouterConnection
	DoneSending bool

	AddressStr string
	}



func (this *TwoWayWindowConnection) waitForSyncAck(sequence uint32){
	syncPacket := this.Router.CreatePacket(router.SYN, sequence, make([]byte, 0))
	this.Router.SendPacket(&syncPacket)

	time.Sleep(time.Second*2)
	if !this.ReceiverWindow.IsSynched {
		go this.waitForSyncAck(sequence)
	}
}


func (this *TwoWayWindowConnection) Connect(){
	startSequence := rand.Uint32()%10


	for i, _ := range this.ReceiverWindow.Buffer{
		this.ReceiverWindow.Buffer[i].SequenceNumber = startSequence+uint32(i)
	}


	go this.waitForSyncAck(startSequence)

	for {
		packet := this.Router.ReadPacket()
		if !this.ReceiverWindow.IsSynched && packet.Type == router.ACK && packet.SequenceNumber == startSequence {
			this.ReceiverWindow.IsSynched = true
			fmt.Printf("Synched client receiver window with sync id %v\n", this.ReceiverWindow.Buffer[0].SequenceNumber)
		}

		if packet.Type == router.SYN && !this.SenderWindow.IsSynched{
			for i, _ := range this.SenderWindow.Buffer{
				this.SenderWindow.Buffer[i].SequenceNumber = packet.SequenceNumber+uint32(i)
			}
			this.SenderWindow.IsSynched = true
			fmt.Printf("Synched client sender window with sync id %v\n", this.SenderWindow.Buffer[0].SequenceNumber)
		}

		if this.SenderWindow.IsSynched && this.ReceiverWindow.IsSynched {
			break
		}


	}

}

func (this *TwoWayWindowConnection) Reset(){
	this.SenderWindow = CreateSenderWindow(this.SenderWindow.Name, this.SenderWindow.BufferSize, this.Router)
	this.ReceiverWindow = CreateReceiverWindow(this.ReceiverWindow.Name, this.ReceiverWindow.BufferSize, this.Router)
	this.DoneSending = false;
}

func (this *TwoWayWindowConnection) GetWaitAndFlushInputBuffer() []byte{
	this.ReceiverWindow.InputBufferMutex.Lock()
	for len(this.ReceiverWindow.InputBuffer) == 0{
		this.ReceiverWindow.InputBufferMutex.Unlock()
		runtime.Gosched()
		this.ReceiverWindow.InputBufferMutex.Lock()
	}
	this.ReceiverWindow.InputBufferMutex.Unlock()

	return this.GetAndFlushInputBuffer()
}

func (this *TwoWayWindowConnection) GetAndFlushInputBuffer() []byte{
	this.ReceiverWindow.InputBufferMutex.Lock()

	data := this.ReceiverWindow.InputBuffer
	this.ReceiverWindow.InputBuffer = make([]byte, 0)

	this.ReceiverWindow.InputBufferMutex.Unlock()

	return data
	}

	func (this *TwoWayWindowConnection) GetLineFromInputBuffer() []byte{
		var data []byte
		this.ReceiverWindow.InputBufferMutex.Lock()
		for {
			if len((this.ReceiverWindow.InputBuffer)) > 0 {
				if this.ReceiverWindow.InputBuffer[0] == '\n' {
					data = append(data, this.ReceiverWindow.InputBuffer[0])
					this.ReceiverWindow.InputBuffer = this.ReceiverWindow.InputBuffer[1:]
					break
				}
			}else{
				this.ReceiverWindow.InputBufferMutex.Unlock()
				runtime.Gosched()
				this.ReceiverWindow.InputBufferMutex.Lock()
			}
		}
		this.ReceiverWindow.InputBufferMutex.Unlock()

		return data
	}

func (this *TwoWayWindowConnection) ProcessPacket(packet *router.Packet){

	switch packet.Type {
	case router.DATA:
		//fmt.Printf("%v - Packet arrived %v %v %v\n",this.ReceiverWindow.Name, packet.Type, packet.SequenceNumber, string(packet.Data))
		this.ReceiverWindow.acceptDataPacket(packet)
		break;
	case router.ACK:
		if(this.endAck > 0){
			if packet.SequenceNumber == this.endAck {
				this.endAck = 0;
			}
		}
		//fmt.Printf("%v - Packet arrived %v %v %v\n",this.SenderWindow.Name, packet.Type, packet.SequenceNumber, string(packet.Data))
		this.SenderWindow.acceptAckPacket(packet)
		break
	case router.NACK:
		//fmt.Printf("%v - Packet arrived %v %v %v\n", this.SenderWindow.Name, packet.Type, packet.SequenceNumber, string(packet.Data))
		this.SenderWindow.acceptNackPacket(packet)
		break
	default:
		break;
		}
	}

func (this *TwoWayWindowConnection) SendPacket(data []byte){
	packet := this.Router.CreatePacket(router.DATA, 0, data)
	this.SenderWindow.enqueueDataPacket(&packet)
	}