package connection

import (
	"encoding/binary"
	"errors"
	"sync"

	openapi "github.com/2nth0nyj/go-trade/exchang/ctrader/connection/proto"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

var protoMessagePool = sync.Pool{
	New: func() interface{} {
		return &openapi.ProtoMessage{}
	},
}

func encode(msg proto.Message) ([]byte, string, error) {
	var payload []byte
	var payloadType openapi.ProtoOAPayloadType

	if m, ok := msg.(*openapi.ProtoOAApplicationAuthReq); ok {
		payloadType = m.GetPayloadType()
		if b, e := proto.Marshal(m); e == nil {
			payload = b
		}
	} else if m, ok := msg.(*openapi.ProtoOAAccountAuthReq); ok {
		payloadType = m.GetPayloadType()
		if b, e := proto.Marshal(m); e == nil {
			payload = b
		}
	} else if m, ok := msg.(*openapi.ProtoOAGetAccountListByAccessTokenReq); ok {
		payloadType = m.GetPayloadType()
		if b, e := proto.Marshal(m); e == nil {
			payload = b
		}
	} else if m, ok := msg.(*openapi.ProtoOATraderReq); ok {
		payloadType = m.GetPayloadType()
		if b, e := proto.Marshal(m); e == nil {
			payload = b
		}
	} else {
		return nil, "", errors.New("not encoded type")
	}

	clientMsgId := uuid.New().String()
	if payload != nil {
		prototMsgType := uint32(payloadType)
		protoMsg := protoMessagePool.Get().(*openapi.ProtoMessage)
		protoMsg.PayloadType = &prototMsgType
		protoMsg.Payload = payload
		protoMsg.ClientMsgId = &clientMsgId
		if b, e := proto.Marshal(protoMsg); e == nil {
			protoMessageLength := len(b)
			protoMessageLengthBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(protoMessageLengthBytes, uint32(protoMessageLength))
			b = append(protoMessageLengthBytes, b...)
			return b, clientMsgId, nil
		}
	}

	return nil, "", errors.New("payload empty")
}
