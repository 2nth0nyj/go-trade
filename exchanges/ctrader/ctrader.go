package ctrader

import (
	"encoding/json"
	"net"

	openapi "github.com/2nth0nyj/go-trade/exchanges/ctrader/proto"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	apiKey string
	secret string
	live   bool
	conn   net.Conn
}

func NewClient(apiKey, secret string, live bool) *Client {
	return &Client{
		apiKey: apiKey,
		secret: secret,
		live:   live,
	}
}

func (c *Client) Start() error {
	var address string
	if c.live {
		address = "live.ctraderapi.com:5036"
	} else {
		address = "demo.ctraderapi.com:5036"
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic("connection to server failed.")
	}
	c.conn = conn
	c.heartbeat()
	return nil
}

func (c *Client) Send(proto proto.Message) error {
	if _, e := json.Marshal(proto); e == nil {
		return e
	} else {
		return e
	}
}

func (c *Client) heartbeat() error {
	payLoadType := openapi.ProtoPayloadType_HEARTBEAT_EVENT
	payload, _ := json.Marshal(&openapi.ProtoHeartbeatEvent{PayloadType: &payLoadType})
	clientMsgId := "heartbeat"
	m := openapi.ProtoMessage{
		PayloadType: proto.Uint32(uint32(payLoadType)),
		Payload:     payload,
		ClientMsgId: &clientMsgId,
	}
	b, _ := json.Marshal(m)
	c.conn.Write(b)
	return nil
}
