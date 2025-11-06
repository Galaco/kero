package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Connectionless packet header (4 bytes of 0xFF)
// These packets are sent before authentication is complete
var ConnectionlessHeader = []byte{0xFF, 0xFF, 0xFF, 0xFF}

// Connectionless packet types (Client -> Server)
const (
	A2S_INFO        byte = 'T' // Request server information
	A2S_PLAYER      byte = 'U' // Request player list
	A2S_RULES       byte = 'V' // Request server rules/cvars
	A2A_PING        byte = 'i' // Ping request
	C2S_CONNECT     byte = 'q' // Initial connection request
	C2S_CONNECT_KEY byte = 'k' // Connection with auth key
)

// Connectionless packet types (Server -> Client)
const (
	S2A_INFO_SRC    byte = 'I' // Server info response (Source engine)
	S2A_INFO_GOLDSRC byte = 'm' // Server info response (GoldSrc)
	S2A_PLAYER      byte = 'D' // Player list response
	S2A_RULES       byte = 'E' // Rules response
	S2C_CHALLENGE   byte = 'A' // Challenge response
	S2C_CONNECTION  byte = 'B' // Connection accepted
	S2A_PING        byte = 'j' // Ping response
)

// ConnectionlessPacket represents a Source Engine connectionless packet
type ConnectionlessPacket struct {
	Type byte   // Packet type identifier
	Data []byte // Packet payload
}

// NewConnectionlessPacket creates a new connectionless packet
func NewConnectionlessPacket(packetType byte, data []byte) *ConnectionlessPacket {
	return &ConnectionlessPacket{
		Type: packetType,
		Data: data,
	}
}

// Serialize converts the packet to bytes for transmission
func (p *ConnectionlessPacket) Serialize() []byte {
	buf := make([]byte, 0, len(ConnectionlessHeader)+1+len(p.Data))
	buf = append(buf, ConnectionlessHeader...)
	buf = append(buf, p.Type)
	buf = append(buf, p.Data...)
	return buf
}

// ParseConnectionless parses a raw packet into a ConnectionlessPacket
func ParseConnectionless(data []byte) (*ConnectionlessPacket, error) {
	if len(data) < 5 {
		return nil, errors.New("packet too short for connectionless format")
	}

	// Verify header (4 bytes of 0xFF)
	if !bytes.Equal(data[0:4], ConnectionlessHeader) {
		return nil, errors.New("invalid connectionless packet header")
	}

	return &ConnectionlessPacket{
		Type: data[4],
		Data: data[5:],
	}, nil
}

// IsConnectionless checks if raw packet data is a connectionless packet
func IsConnectionless(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return bytes.Equal(data[0:4], ConnectionlessHeader)
}

// String returns a human-readable representation of the packet
func (p *ConnectionlessPacket) String() string {
	typeName := "Unknown"
	switch p.Type {
	case A2S_INFO:
		typeName = "A2S_INFO"
	case A2S_PLAYER:
		typeName = "A2S_PLAYER"
	case A2S_RULES:
		typeName = "A2S_RULES"
	case A2A_PING:
		typeName = "A2A_PING"
	case C2S_CONNECT:
		typeName = "C2S_CONNECT"
	case C2S_CONNECT_KEY:
		typeName = "C2S_CONNECT_KEY"
	case S2A_INFO_SRC:
		typeName = "S2A_INFO_SRC"
	case S2A_INFO_GOLDSRC:
		typeName = "S2A_INFO_GOLDSRC"
	case S2A_PLAYER:
		typeName = "S2A_PLAYER"
	case S2A_RULES:
		typeName = "S2A_RULES"
	case S2C_CHALLENGE:
		typeName = "S2C_CHALLENGE"
	case S2C_CONNECTION:
		typeName = "S2C_CONNECTION"
	case S2A_PING:
		typeName = "S2A_PING"
	}
	return fmt.Sprintf("ConnectionlessPacket{Type: %s (0x%02X), DataLen: %d}", typeName, p.Type, len(p.Data))
}

// PacketReader helps parse packet data with proper error handling
type PacketReader struct {
	data   []byte
	offset int
}

// NewPacketReader creates a new packet reader
func NewPacketReader(data []byte) *PacketReader {
	return &PacketReader{
		data:   data,
		offset: 0,
	}
}

// ReadByte reads a single byte
func (r *PacketReader) ReadByte() (byte, error) {
	if r.offset >= len(r.data) {
		return 0, errors.New("packet reader: end of buffer")
	}
	b := r.data[r.offset]
	r.offset++
	return b, nil
}

// ReadBytes reads n bytes
func (r *PacketReader) ReadBytes(n int) ([]byte, error) {
	if r.offset+n > len(r.data) {
		return nil, errors.New("packet reader: not enough bytes")
	}
	bytes := r.data[r.offset : r.offset+n]
	r.offset += n
	return bytes, nil
}

// ReadString reads a null-terminated string
func (r *PacketReader) ReadString() (string, error) {
	start := r.offset
	for r.offset < len(r.data) {
		if r.data[r.offset] == 0 {
			str := string(r.data[start:r.offset])
			r.offset++ // Skip null terminator
			return str, nil
		}
		r.offset++
	}
	return "", errors.New("packet reader: unterminated string")
}

// ReadInt8 reads a signed 8-bit integer
func (r *PacketReader) ReadInt8() (int8, error) {
	b, err := r.ReadByte()
	return int8(b), err
}

// ReadUint16 reads an unsigned 16-bit integer (little endian)
func (r *PacketReader) ReadUint16() (uint16, error) {
	bytes, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(bytes), nil
}

// ReadInt16 reads a signed 16-bit integer (little endian)
func (r *PacketReader) ReadInt16() (int16, error) {
	u, err := r.ReadUint16()
	return int16(u), err
}

// ReadUint32 reads an unsigned 32-bit integer (little endian)
func (r *PacketReader) ReadUint32() (uint32, error) {
	bytes, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(bytes), nil
}

// ReadInt32 reads a signed 32-bit integer (little endian)
func (r *PacketReader) ReadInt32() (int32, error) {
	u, err := r.ReadUint32()
	return int32(u), err
}

// ReadFloat32 reads a 32-bit float (little endian)
func (r *PacketReader) ReadFloat32() (float32, error) {
	bytes, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint32(bytes)
	return float32(bits), nil
}

// Remaining returns the number of unread bytes
func (r *PacketReader) Remaining() int {
	return len(r.data) - r.offset
}

// HasMore returns true if there are more bytes to read
func (r *PacketReader) HasMore() bool {
	return r.offset < len(r.data)
}

// PacketWriter helps build packet data
type PacketWriter struct {
	buf *bytes.Buffer
}

// NewPacketWriter creates a new packet writer
func NewPacketWriter() *PacketWriter {
	return &PacketWriter{
		buf: &bytes.Buffer{},
	}
}

// WriteByte writes a single byte
func (w *PacketWriter) WriteByte(b byte) {
	w.buf.WriteByte(b)
}

// WriteBytes writes a byte slice
func (w *PacketWriter) WriteBytes(data []byte) {
	w.buf.Write(data)
}

// WriteString writes a null-terminated string
func (w *PacketWriter) WriteString(s string) {
	w.buf.WriteString(s)
	w.buf.WriteByte(0)
}

// WriteInt8 writes a signed 8-bit integer
func (w *PacketWriter) WriteInt8(v int8) {
	w.buf.WriteByte(byte(v))
}

// WriteUint16 writes an unsigned 16-bit integer (little endian)
func (w *PacketWriter) WriteUint16(v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	w.buf.Write(b[:])
}

// WriteInt16 writes a signed 16-bit integer (little endian)
func (w *PacketWriter) WriteInt16(v int16) {
	w.WriteUint16(uint16(v))
}

// WriteUint32 writes an unsigned 32-bit integer (little endian)
func (w *PacketWriter) WriteUint32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf.Write(b[:])
}

// WriteInt32 writes a signed 32-bit integer (little endian)
func (w *PacketWriter) WriteInt32(v int32) {
	w.WriteUint32(uint32(v))
}

// WriteFloat32 writes a 32-bit float (little endian)
func (w *PacketWriter) WriteFloat32(v float32) {
	w.WriteUint32(uint32(v))
}

// Bytes returns the written bytes
func (w *PacketWriter) Bytes() []byte {
	return w.buf.Bytes()
}
