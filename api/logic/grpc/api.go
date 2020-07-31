package grpc

import (
	"fmt"
)

type MidType string

func (mid MidType) IsNotZero() bool {
	return mid != ""
}

func (m *ConnectReq) XString() string {
	return fmt.Sprintf("server:%s token:%s", m.Server, m.Token)
}

func (m *OnlineReq) XString() string {
	return fmt.Sprintf("server:%s", m.Server)
}

func (m *OnlineReply) XString() string {
	return fmt.Sprintf("rooms:%d", len(m.AllRoomCount))
}
