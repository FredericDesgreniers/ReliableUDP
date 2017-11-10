package reliable

type Packet struct{
	Type uint8
	SequenceNumber uint32

	Address []byte
	Port uint16

	Data []byte
}

func CreatePacket(Type uint8, SequenceNumber uint32, Address []byte, Port uint16, Data []byte) Packet{
	return Packet{Type, SequenceNumber, Address, Port, Data}
}