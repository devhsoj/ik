package ik

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const ProtoVersion byte = 1

func craftPacketMetadata(version byte, commandName string, dataLength int) []byte {
	buf := make([]byte, 1+2+len(commandName)+4)

	buf[0] = version

	binary.LittleEndian.PutUint16(buf[1:], uint16(len(commandName)))

	copy(buf[1+2:], commandName)

	binary.PutVarint(buf[1+2+len(commandName):], int64(dataLength))

	return buf
}

func readPacketMetadata(r *bufio.Reader) (version byte, commandName string, dataLength int, err error) {
	commandNameLengthBuf := make([]byte, 2)
	dataLengthBuf := make([]byte, 4)

	var commandNameBuf []byte
	var dataLength_ int64

	version, err = r.ReadByte()

	if err != nil {
		return 0, "", 0, err
	}

	if version != ProtoVersion {
		return 0, "", 0, fmt.Errorf("invalid packet version: %d", version)
	}

	if _, err = io.ReadFull(r, commandNameLengthBuf); err != nil {
		return 0, "", 0, err
	}

	commandNameBuf = make([]byte, binary.LittleEndian.Uint16(commandNameLengthBuf))

	if _, err = io.ReadFull(r, commandNameBuf); err != nil {
		return 0, "", 0, err
	}

	if _, err = io.ReadFull(r, dataLengthBuf); err != nil {
		return 0, "", 0, err
	}

	dataLength_, _ = binary.Varint(dataLengthBuf)

	return version, string(commandNameBuf), int(dataLength_), nil
}
