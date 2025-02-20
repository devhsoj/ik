package ik

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const ProtoVersion byte = 1

func craftPacketMetadata(version byte, event string, dataLength int) []byte {
	buf := make([]byte, 1+2+len(event)+4)

	buf[0] = version

	binary.LittleEndian.PutUint16(buf[1:], uint16(len(event)))

	copy(buf[1+2:], event)

	binary.PutVarint(buf[1+2+len(event):], int64(dataLength))

	return buf
}

func readPacketMetadata(r *bufio.Reader) (version byte, event string, dataLength int, err error) {
	eventLengthBuf := make([]byte, 2)
	dataLengthBuf := make([]byte, 4)

	var eventBuf []byte
	var dataLength_ int64

	version, err = r.ReadByte()

	if err != nil {
		return 0, "", 0, err
	}

	if version != ProtoVersion {
		return 0, "", 0, fmt.Errorf("ik: invalid version in packet metadata: %d", version)
	}

	if _, err = io.ReadFull(r, eventLengthBuf); err != nil {
		return 0, "", 0, err
	}

	eventBuf = make([]byte, binary.LittleEndian.Uint16(eventLengthBuf))

	if _, err = io.ReadFull(r, eventBuf); err != nil {
		return 0, "", 0, err
	}

	if _, err = io.ReadFull(r, dataLengthBuf); err != nil {
		return 0, "", 0, err
	}

	dataLength_, _ = binary.Varint(dataLengthBuf)

	return version, string(eventBuf), int(dataLength_), nil
}
