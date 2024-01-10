package connection

// +--------------------------+-----------------------------------------+
// | Message Length (4 bytes) | Serialized ProtoMessage object (byte[]) |
// +--------------------------+-----------------------------------------+
// 						   |<---------- Message Length ------------->|

// +----------------------+
// | int32 payloadType    |
// | byte[] payload       |
// | string clientMsgId   |
// +----------------------+

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	openapi "github.com/2nth0nyj/go-trade/exchang/ctrader/connection/proto"
	"google.golang.org/protobuf/proto"
)

type Connection struct {
	net.Conn
	connected         bool
	accountAuthorized bool
	live              bool
	readChan          chan proto.Message
	writeChan1        chan proto.Message
	writeChan2        chan []byte
	closeChan         chan struct{}
	waitingResponse   sync.Map
	clientId          string
	clientSecret      string
	appAuthorized     bool
	accessToken       string
	ctid              int64
}

func NewConnection(clientId, clientSecret, accessToken string, ctid int64, live bool) *Connection {
	c := &Connection{
		clientId:        clientId,
		clientSecret:    clientSecret,
		live:            live,
		readChan:        make(chan proto.Message, 128),
		writeChan1:      make(chan proto.Message, 128),
		writeChan2:      make(chan []byte, 128),
		closeChan:       make(chan struct{}),
		waitingResponse: sync.Map{},
		accessToken:     accessToken,
		ctid:            ctid,
	}
	go c.login()
	return c
}

func (c *Connection) Stop() {
	c.closeChan <- struct{}{}
}

func (c *Connection) login() {
	host := "demo.ctraderapi.com:5035"
	if c.live {
		host = "live.ctraderapi.com:5035"
	}
	if conn, err := tls.Dial("tcp", host, nil); err == nil {
		c.Conn = conn
		c.connected = true
		c.accountAuthorized = false

		go c.read()
		go c.write()
		go c.heartbeat()

		authReq := &openapi.ProtoOAApplicationAuthReq{
			ClientId:     &c.clientId,
			ClientSecret: &c.clientSecret,
		}
		authRes := c.SendProtoAndWaitResponse(authReq)
		if _, ok := authRes.(*openapi.ProtoOAApplicationAuthRes); ok {
			c.appAuthorized = true
		}

		accountAuthRes := c.SendProtoAndWaitResponse(&openapi.ProtoOAAccountAuthReq{CtidTraderAccountId: &c.ctid, AccessToken: &c.accessToken})
		if _, ok := accountAuthRes.(*openapi.ProtoOAAccountAuthRes); ok {
			c.accountAuthorized = true
		}

		<-c.closeChan
	}
}

func (c *Connection) SendProtoAndWaitResponse(m proto.Message) proto.Message {
	if c.connected {
		var v proto.Message = nil
		if writingBytes, clientMsgId, err := encode(m); err == nil {
			clientWaitChannel := make(chan proto.Message)
			c.waitingResponse.Store(clientMsgId, clientWaitChannel)
			c.writeChan2 <- writingBytes
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
			defer cancelFunc()
			select {
			case <-ctx.Done():
				fmt.Printf("cancel done...")
				cancelFunc()
				close(c.closeChan)
				v = nil
			case returnedMsg := <-clientWaitChannel:
				v = returnedMsg
			}
			close(clientWaitChannel)
		}
		return v
	}
	return nil
}

func (c *Connection) write() {
	for {
		select {
		case msg := <-c.writeChan1:
			b, _, _ := encode(msg)
			c.Write(b)
		case b := <-c.writeChan2:
			c.Write(b)
		case <-c.closeChan:
			return
		}
	}
}

func (c *Connection) read() {
	for {
		select {
		case <-c.closeChan:
			return
		default:
			b := make([]byte, 4)
			if _, err := io.ReadFull(c.Conn, b); err == nil {
				restLength := binary.BigEndian.Uint32(b)
				b := make([]byte, restLength)
				if _, err := io.ReadFull(c.Conn, b); err == nil {
					m := openapi.ProtoMessage{}
					if proto.Unmarshal(b, &m) == nil && m.PayloadType != nil && m.Payload != nil && m.ClientMsgId != nil {
						payloadType := (openapi.ProtoPayloadType)(*(m.PayloadType))
						payload := m.Payload
						clientMsgId := *(m.ClientMsgId)
						responseChannel, ok := c.waitingResponse.LoadAndDelete(clientMsgId)
						switch payloadType {
						case openapi.ProtoPayloadType_HEARTBEAT_EVENT:
							fmt.Printf("Received heartbeat. \n")
						case openapi.ProtoPayloadType(openapi.ProtoOAPayloadType_PROTO_OA_ERROR_RES):
							protoErrorMessage := openapi.ProtoErrorRes{}
							if proto.Unmarshal(b, &protoErrorMessage) == nil {
								fmt.Printf("protoErrorMsg: %v\n", *(protoErrorMessage.ErrorCode))
							}
						case openapi.ProtoPayloadType(openapi.ProtoOAPayloadType_PROTO_OA_APPLICATION_AUTH_RES):
							protoOAApplicationAuthRes := openapi.ProtoOAApplicationAuthRes{}
							if proto.Unmarshal(payload, &protoOAApplicationAuthRes) == nil && ok {
								if rc, ok := responseChannel.(chan proto.Message); ok {
									rc <- &protoOAApplicationAuthRes
								}
							}
						case openapi.ProtoPayloadType(openapi.ProtoOAPayloadType_PROTO_OA_ACCOUNT_AUTH_RES):
							protoOAAccountAuthRes := openapi.ProtoOAAccountAuthRes{}
							if proto.Unmarshal(payload, &protoOAAccountAuthRes) == nil && ok {
								if rc, ok := responseChannel.(chan proto.Message); ok {
									rc <- &protoOAAccountAuthRes
								}
							}
						case openapi.ProtoPayloadType(openapi.ProtoOAPayloadType_PROTO_OA_GET_ACCOUNTS_BY_ACCESS_TOKEN_RES):
							res := openapi.ProtoOAGetAccountListByAccessTokenRes{}
							if proto.Unmarshal(payload, &res) == nil && ok {
								if rc, ok := responseChannel.(chan proto.Message); ok {
									rc <- &res
								}
							}
						case openapi.ProtoPayloadType(openapi.ProtoOAPayloadType_PROTO_OA_TRADER_RES):
							res := openapi.ProtoOATraderRes{}
							if proto.Unmarshal(payload, &res) == nil && ok {
								if rc, ok := responseChannel.(chan proto.Message); ok {
									rc <- &res
								}
							}
						default:
							fmt.Printf("Unhandled type: %v", payloadType)
						}
					}
				}
			}
		}
	}
}

func (c *Connection) heartbeat() {
	f := func() interface{} {
		heartbeatProtocol := openapi.ProtoHeartbeatEvent{}
		return &heartbeatProtocol
	}
	heartBeatPool := sync.Pool{New: f}
	for {
		select {
		case <-c.closeChan:
			return
		default:
			h := heartBeatPool.Get().(*openapi.ProtoHeartbeatEvent)
			b, _, _ := encode(h)
			heartBeatPool.Put(h)
			c.writeChan2 <- b
		}
		time.Sleep(time.Duration(10 * time.Second))
	}
}
