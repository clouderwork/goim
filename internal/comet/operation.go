package comet

import (
	"context"
	"time"

	model "github.com/Terry-Mao/goim/api/comet/grpc"
	logicapi "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/pkg/strings"
	log "github.com/golang/glog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// Connect connected a connection.
func (s *Server) Connect(c context.Context, p *model.Proto, cookie string) (mid logicapi.MidType, key string, platform string, rid string, accepts []int32, heartbeat time.Duration, err error) {
	reply, err := s.rpcClient.Connect(c, &logicapi.ConnectReq{
		Server: s.serverID,
		Cookie: cookie,
		Token:  p.Body,
	})
	if err != nil {
		return
	}
	return logicapi.MidType(reply.Mid), reply.Key, reply.Platform, reply.RoomID, reply.Accepts, time.Duration(reply.Heartbeat), nil
}

// Disconnect disconnected a connection.
func (s *Server) Disconnect(c context.Context, mid logicapi.MidType, key string) (err error) {
	_, err = s.rpcClient.Disconnect(context.Background(), &logicapi.DisconnectReq{
		Server: s.serverID,
		Mid:    string(mid),
		Key:    key,
	})
	return
}

// Heartbeat heartbeat a connection session.
func (s *Server) Heartbeat(ctx context.Context, mid logicapi.MidType, key string) (err error) {
	_, err = s.rpcClient.Heartbeat(ctx, &logicapi.HeartbeatReq{
		Server: s.serverID,
		Mid:    string(mid),
		Key:    key,
	})
	return
}

// RenewOnline renew room online.
func (s *Server) RenewOnline(ctx context.Context, serverID string, rommCount map[string]int32) (allRoom map[string]int32, err error) {
	reply, err := s.rpcClient.RenewOnline(ctx, &logicapi.OnlineReq{
		Server:    s.serverID,
		RoomCount: rommCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

// Receive receive a message.
func (s *Server) Receive(ctx context.Context, deviceID string, mid logicapi.MidType, platform string, p *model.Proto) (err error) {
	_, err = s.rpcClient.Receive(ctx, &logicapi.ReceiveReq{
		Mid:      string(mid),
		DeviceId: deviceID,
		Platform: platform,
		Proto:    p,
	})
	return
}

// Operate operate.
func (s *Server) Operate(ctx context.Context, p *model.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	case model.OpChangeRoom:
		if err := b.ChangeRoom(string(p.Body), ch); err != nil {
			log.Errorf("b.ChangeRoom(%s) error(%v)", p.Body, err)
		}
		if err := s.Receive(ctx, ch.Key, ch.Mid, ch.Platform, p); err != nil {
			log.Errorf("s.Report(%d) op:%d error(%v)", ch.Mid, p.Op, err)
		}
		p.Op = model.OpChangeRoomReply
	case model.OpSub:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.Watch(ops...)
		}
		p.Op = model.OpSubReply
	case model.OpUnsub:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.UnWatch(ops...)
		}
		p.Op = model.OpUnsubReply
	default:
		// TODO ack ok&failed
		if err := s.Receive(ctx, ch.Key, ch.Mid, ch.Platform, p); err != nil {
			log.Errorf("s.Report(%d) op:%d error(%v)", ch.Mid, p.Op, err)
		}
		p.Body = nil
	}
	return nil
}
