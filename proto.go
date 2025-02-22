package ik

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const ProtoVersion byte = 2

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

func sendPacket(w *bufio.Writer, version byte, event string, data []byte) error {
	buf := craftPacketMetadata(version, event, len(data))

	if _, err := w.Write(buf); err != nil {
		return nil
	}

	if _, err := w.Write(data); err != nil {
		return nil
	}

	return w.Flush()
}
