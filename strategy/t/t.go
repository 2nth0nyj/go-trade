package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	openapi "github.com/2nth0nyj/go-trade/exchang/ctrader/connection/proto"
	"google.golang.org/protobuf/proto"
)

func main() {
	args := os.Args
	fmt.Printf("args = %v \n", args[1])
	if args[1] == "s" {
		s()
	}
	if args[1] == "c" {
		c()
	}
	if args[1] == "h" {
		heartbeatEnocded()
	}
	if args[1] == "d" {
		decode()
	}
}

func s() {
	listener, _ := net.Listen("tcp", "127.0.0.1:30000")
	fmt.Printf("listenning.on  %v\n", listener)
	conn, _ := listener.Accept()
	b := make([]byte, 4)
	n, e := io.ReadFull(conn, b)
	fmt.Printf("n = %v, e = %v \n", n, e)
}

func c() {
	conn, _ := net.Dial("tcp", "127.0.0.1:30000")
	for {
		n, err := conn.Write([]byte{0x1})
		fmt.Printf("n = %v, err = %v\n", n, err)
		time.Sleep(time.Duration(5 * time.Second))
	}
}

func heartbeatEnocded() {
	payloadType := openapi.ProtoPayloadType_HEARTBEAT_EVENT
	heartbeat := openapi.ProtoHeartbeatEvent{PayloadType: &payloadType}
	p, e := proto.Marshal(&heartbeat)
	fmt.Printf("%v, %v \n", p, e)
	fmt.Printf("length = %v", len(p))
	b := make([]byte, 4)
	l := len(p)
	binary.LittleEndian.PutUint32(b, uint32(l))
	b = append(b, p...)
	fmt.Printf("b = %v\n", b)
}

func decode() {
	fmt.Printf("test decode.")
	b := []byte{8, 180, 16, 18, 112, 8, 180, 16, 18, 55, 50, 49, 56, 51, 95, 121, 89, 89, 119, 99, 86, 121, 119, 52, 48, 76, 77, 112, 108, 108, 116, 71, 80, 67, 102, 98, 115, 100, 66, 97, 106, 100, 73, 67, 102, 100, 114, 78, 102, 83, 69, 72, 83, 89, 70, 107, 100, 70, 98, 83, 71, 52, 117, 119, 49, 26, 50, 104, 72, 116, 50, 70, 120, 119, 83, 66, 89, 48, 54, 98, 78, 48, 98, 72, 98, 73, 48, 69, 83, 111, 102, 114, 76, 121, 121, 67, 87, 76, 104, 97, 48, 119, 114, 51, 71, 118, 56, 102, 88, 79, 66, 100, 119, 108, 99, 51, 115, 26, 36, 49, 99, 53, 50, 56, 52, 97, 56, 45, 50, 98, 50, 51, 45, 52, 99, 53, 53, 45, 98, 54, 57, 55, 45, 50, 99, 55, 97, 56, 54, 55, 55, 57, 55, 101, 97}
	v := openapi.ProtoMessage{}
	proto.Unmarshal(b, &v)
	fmt.Printf("v = %v, payloadType = %v, payload = %v\n", *(v.ClientMsgId), *(v.PayloadType), v.Payload)
	w := openapi.ProtoOAApplicationAuthReq{}
	proto.Unmarshal(v.Payload, &w)
	fmt.Printf("type = %d clientId = %v clientSecret = %v\n", *(w.PayloadType), *(w.ClientId), *(w.ClientSecret))
}
