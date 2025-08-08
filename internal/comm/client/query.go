// Package client provides simple functions to communicate with the socket.
package client

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/abenz1267/elephant/internal/comm/pb/pb"
	"google.golang.org/protobuf/proto"
)

func Query(data string) {
	v := strings.Split(data, ";")
	maxresults, _ := strconv.Atoi(v[2])

	req := pb.QueryRequest{
		Providers:  strings.Split(v[0], ","),
		Query:      v[1],
		Maxresults: int32(maxresults),
	}

	b, err := proto.Marshal(&req)
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var buffer bytes.Buffer
	buffer.Write([]byte{0})

	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(b)))
	buffer.Write(lengthBuf)
	buffer.Write(b)

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(conn)

	for {
		header, err := reader.Peek(5)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if header[0] == done {
			break
		}

		if header[0] != 0 {
			panic("invalid protocol prefix")
		}

		length := binary.BigEndian.Uint32(header[1:5])

		msg := make([]byte, 5+length)
		_, err = io.ReadFull(reader, msg)
		if err != nil {
			panic(err)
		}

		payload := msg[5:]

		resp := &pb.QueryResponse{}
		if err := proto.Unmarshal(payload, resp); err != nil {
			panic(err)
		}

		fmt.Println(resp)
	}
}
