package reliable

import (
"../router"
	"time"
	"sync"
)




type SenderWindow struct{
	Name string

	BufferSize uint32
	Buffer []WindowSlot
	AcceptedQueue []*router.Packet

	routerConnection *router.RouterConnection

	mutex sync.Mutex

	IsSynched bool
}


type ReceiverWindow struct{
	Name string

	InputBuffer []byte
	InputBufferMutex sync.Mutex


	BufferSize uint32
	Buffer []WindowSlot
	AcceptedQueue []*router.Packet

	routerConnection *router.RouterConnection

	mutex sync.Mutex

	IsSynched bool
}


type WindowSlot struct{
	SequenceNumber uint32
	Packet *router.Packet
	canBeDiscarded bool
	sent bool
}

func (this *SenderWindow) enqueueDataPacket(packet *router.Packet){
	for i:=uint32(0); int(i) < len(this.Buffer); i++{
		if this.Buffer[i].Packet == nil {
			this.Buffer[i].Packet = packet
			packet.SequenceNumber = this.Buffer[i].SequenceNumber

			if i < this.BufferSize{
				this.sendDataPacket(&this.Buffer[i])
			}

			return
		}
	}

	this.Buffer = append(this.Buffer, WindowSlot{SequenceNumber:this.Buffer[this.BufferSize-1].SequenceNumber+1})
}

func (this *SenderWindow) sendDataPacket(slot *WindowSlot){

	this.routerConnection.SendPacket(slot.Packet)
	slot.sent = true
	go func(){
		time.Sleep(time.Second*5)
		if !slot.canBeDiscarded && !(slot.SequenceNumber < this.Buffer[0].SequenceNumber){
			//fmt.Println(this.Name, " Lost packet "+strconv.Itoa(int(slot.SequenceNumber)))
			this.sendDataPacket(slot)
		}else{

		}
	}()
}

func (this *SenderWindow) acceptAckPacket(packet *router.Packet){
	//fmt.Println(this.Name, " ACK for ", packet.SequenceNumber)
	if(packet.SequenceNumber < this.Buffer[0].SequenceNumber){
		return
	}
	for i, slot := range this.Buffer{
		if slot.SequenceNumber == packet.SequenceNumber {
			this.Buffer[i].canBeDiscarded = true
			break
		}
	}
	this.attemptSlide()

	for i :=uint32(0); i < this.BufferSize; i++{
		if(!this.Buffer[i].sent && this.Buffer[i].Packet != nil){
			this.sendDataPacket(&this.Buffer[i])
		}
	}


}

func (this *SenderWindow) acceptNackPacket(packet *router.Packet){
	//fmt.Println(this.Name, " ->  NACK  <- for ", packet.SequenceNumber)

	for index, slot := range this.Buffer{
		if slot.SequenceNumber == packet.SequenceNumber {
			this.sendDataPacket(&this.Buffer[index])
			break
		}
	}
}


func (this *ReceiverWindow) acceptDataPacket(packet *router.Packet){
	foundPlace := false
	if(packet.SequenceNumber < this.Buffer[0].SequenceNumber){
		ackPacket := this.routerConnection.CreatePacket(router.ACK, packet.SequenceNumber, []byte(""))
		this.routerConnection.SendPacket(&ackPacket)
		return
	}
	for i := uint32(0); i < this.BufferSize; i++{

		slot := &this.Buffer[i]
		if slot.SequenceNumber == packet.SequenceNumber{
			slot.canBeDiscarded = true
			slot.Packet = packet

			ackPacket := this.routerConnection.CreatePacket(router.ACK, slot.SequenceNumber, []byte(""))

			this.routerConnection.SendPacket(&ackPacket)

			//fmt.Println("Packet arrived! ", slot.SequenceNumber)

			foundPlace = true
			break;
		}
	}

	if !foundPlace{
		//fmt.Println(this.Name, " Packet ", packet.SequenceNumber, " somehow not in buffer")
	}

	this.attemptSlide()
}
func (this *ReceiverWindow) attemptSlide(){
	this.mutex.Lock()
	slideFrom := -1
	foundData := false

	for i := this.BufferSize-1; i >= 0 && i < this.BufferSize; i--{
		if this.Buffer[i].canBeDiscarded {
			if slideFrom < 0 {
				slideFrom = int(i)
				foundData = true
			}
		} else {
			if foundData{
				nack := this.routerConnection.CreatePacket(router.NACK, this.Buffer[i].SequenceNumber, []byte{})
				this.routerConnection.SendPacket(&nack)
				//fmt.Println(this.Name, " Sending nack for ", this.Buffer[i].SequenceNumber)
			}
			slideFrom = -1
		}
	}

	if slideFrom >= 0 {
		for i2:=0; i2 <= slideFrom; i2++{
			this.InputBufferMutex.Lock();
			this.InputBuffer = append(this.InputBuffer, this.Buffer[0].Packet.Data...);
			this.InputBufferMutex.Unlock();

			this.Buffer = this.Buffer[1:]
			this.Buffer = append(this.Buffer, WindowSlot{SequenceNumber:this.Buffer[this.BufferSize-2].SequenceNumber+1})
		}

	}else{
	}

	//fmt.Println("Receiver window is now from ", this.Buffer[0].SequenceNumber, " to ", this.Buffer[this.BufferSize-1].SequenceNumber)
	this.mutex.Unlock()
}
func (this *SenderWindow) attemptSlide(){
	slideFrom := -1


	for i := this.BufferSize-1; i >= 0 && i < this.BufferSize; i--{
		if this.Buffer[i].canBeDiscarded {
			if slideFrom < 0 {
				slideFrom = int(i)

			}
		} else {
			slideFrom = -1
		}
	}

	if slideFrom >= 0 {
		for i2:=0; i2 <= slideFrom; i2++{
			this.Buffer = this.Buffer[1:]
			this.Buffer = append(this.Buffer, WindowSlot{SequenceNumber:this.Buffer[this.BufferSize-2].SequenceNumber+1})
		}

	}else{
	}

	//fmt.Println("Sender indow is now from ", this.Buffer[0].SequenceNumber, " to ", this.Buffer[this.BufferSize-1].SequenceNumber)
}

