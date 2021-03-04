package ProtobufClient

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type SubscriptionType int

const (
	SubscriptionTypeUnknown SubscriptionType = iota
	SubscriptionTypeStream
	SubscriptionTypeTrade
)

// DeserializeClientMessage decodes the client message and switches based on the message format.
// This is needed because older clients may be pushing ClientIdentificationMessages not wrapped
// in the ClientMessage envelope.
func DeserializeClientMessage(msgData []byte, sf SerializationFormat) (msg ClientMessage, err error) {
	switch sf {
	case ProtobufSerialization:
		// Unmarshal client message envelope
		err = proto.Unmarshal(msgData, &msg)
		if err != nil {
			var identificationMsg ClientIdentificationMessage
			err = proto.Unmarshal(msgData, &identificationMsg)
			msg = ClientMessage{
				Body: &ClientMessage_Identification{
					Identification: &identificationMsg,
				},
			}
		}
	case JSONSerialization:
		err = jsonpb.Unmarshal(bytes.NewBuffer(msgData), &msg)
	}
	return msg, err
}

// SubsFromString converts a given list of legacy string-based keys to list ClientSubscription.
func SubsFromString(kind SubscriptionType, keys []string) []*ClientSubscription {
	ret := make([]*ClientSubscription, 0, len(keys))

	for _, v := range keys {
		ret = append(ret, SubFromString(kind, v))
	}

	return ret
}

// KeysFromSubs converts a given list of ClientSubscription to list of legacy string-based keys.
func KeysFromSubs(subs []*ClientSubscription) []string {
	ret := make([]string, 0, len(subs))

	for _, v := range subs {
		ret = append(ret, KeyFromSub(v))
	}

	return ret
}

// SubFromString converts a given legacy string-based key to ClientSubscription.
func SubFromString(kind SubscriptionType, key string) *ClientSubscription {
	switch kind {
	case SubscriptionTypeStream:
		return &ClientSubscription{
			Body: &ClientSubscription_StreamSubscription{
				StreamSubscription: &StreamSubscription{
					Resource: key,
				},
			},
		}
	case SubscriptionTypeTrade:
		return &ClientSubscription{
			Body: &ClientSubscription_TradeSubscription{
				TradeSubscription: &TradeSubscription{
					// TODO include api key here?
					MarketId: key,
				},
			},
		}
	}

	// Should not happen.
	panic("SubFromString: invalid state")
}

// KeyFromString converts a given ClientSubscription to legacy string-based key.
func KeyFromSub(sub *ClientSubscription) string {
	if sub == nil {
		return ""
	}
	switch v := sub.Body.(type) {
	case *ClientSubscription_StreamSubscription:
		return v.StreamSubscription.GetResource()
	case *ClientSubscription_TradeSubscription:
		return v.TradeSubscription.GetMarketId()
	default:
		return ""
	}
}
