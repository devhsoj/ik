package ik

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

/*
protocol format:
<VERSION: 1b><EVENT NAME LENGTH: 1b><EVENT NAME: Xb><DATA LENGTH: 4b><DATA Xb>
*/

// ProtoVersion is the current version flag for the ik protocol.
const ProtoVersion byte = 2

// MaxDataLength is the maximum length of data allowed to be sent in one packet. ~ 2 GiB.
const MaxDataLength int = 1<<31 - 1

var ErrDataTooLarge = errors.New("data too large, must be less than 2 GiB")

// craftPacketMetadata returns a buffer of metadata regarding a packet to be sent. This buffer contains a version byte,
// an event name length byte, an event name, and the length of data to be sent.
func craftPacketMetadata(version byte, event string, dataLength int) []byte {
	if len(event) > 255 {
		event = event[:255]
	}

	buf := make([]byte, 1+1+len(event)+4)

	buf[0] = version
	buf[1] = uint8(len(event))

	copy(buf[1+1:], event)

	binary.PutVarint(buf[1+1+len(event):], int64(dataLength))

	return buf
}

// readPacketMetadata reads metadata in the format returned from craftPacketMetadata from the passed reader.
func readPacketMetadata(r *bufio.Reader) (version byte, event string, dataLength int, err error) {
	dataLengthBuf := make([]byte, 4)

	var eventLength byte
	var eventBuf []byte
	var dataLength_ int64

	version, err = r.ReadByte()

	if err != nil {
		return 0, "", 0, err
	}

	if version != ProtoVersion {
		return 0, "", 0, fmt.Errorf("ik: invalid version in packet metadata: %d", version)
	}

	eventLength, err = r.ReadByte()

	if err != nil {
		return 0, "", 0, err
	}

	eventBuf = make([]byte, eventLength)

	if _, err = io.ReadFull(r, eventBuf); err != nil {
		return 0, "", 0, err
	}

	if _, err = io.ReadFull(r, dataLengthBuf); err != nil {
		return 0, "", 0, err
	}

	dataLength_, _ = binary.Varint(dataLengthBuf)

	return version, string(eventBuf), int(dataLength_), nil
}

// sendPacket writes metadata from the given version, event, and data in the format returned from craftPacketMetadata
// to the writer, then writes the passed data buf to the writer while flushing to assure it is 'sent'.
func sendPacket(w *bufio.Writer, version byte, event string, data []byte) error {
	if len(data) > MaxDataLength {
		return ErrDataTooLarge
	}

	buf := craftPacketMetadata(version, event, len(data))

	if _, err := w.Write(buf); err != nil {
		return nil
	}

	if _, err := w.Write(data); err != nil {
		return nil
	}

	return w.Flush()
}
