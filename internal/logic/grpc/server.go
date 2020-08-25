package grpc

import (
	"context"
	"net"
	"time"

	logicapi "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/logic"
	"github.com/Terry-Mao/goim/internal/logic/conf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// use gzip decoder
	_ "google.golang.org/grpc/encoding/gzip"
)

// New logic grpc server
func New(c *conf.RPCServer, l *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	logicapi.RegisterLogicServer(srv, &server{l})
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *logic.Logic
}

var _ logicapi.LogicServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *logicapi.PingReq) (*logicapi.PingReply, error) {
	return &logicapi.PingReply{}, nil
}

// Close Service
func (s *server) Close(ctx context.Context, req *logicapi.CloseReq) (*logicapi.CloseReply, error) {
	return &logicapi.CloseReply{}, nil
}

// Connect connect a conn.
func (s *server) Connect(ctx context.Context, req *logicapi.ConnectReq) (*logicapi.ConnectReply, error) {
	mid, key, room, platform, accepts, hb, err := s.srv.Connect(ctx, req.Server, req.Cookie, req.Token)
	if err != nil {
		return &logicapi.ConnectReply{}, err
	}
	return &logicapi.ConnectReply{Mid: string(mid), Key: key, RoomID: room, Platform: platform, Accepts: accepts, Heartbeat: hb}, nil
}

// Disconnect disconnect a conn.
func (s *server) Disconnect(ctx context.Context, req *logicapi.DisconnectReq) (*logicapi.DisconnectReply, error) {
	has, err := s.srv.Disconnect(ctx, logicapi.MidType(req.Mid), req.Key, req.Server)
	if err != nil {
		return &logicapi.DisconnectReply{}, err
	}
	return &logicapi.DisconnectReply{Has: has}, nil
}

// Heartbeat beartbeat a conn.
func (s *server) Heartbeat(ctx context.Context, req *logicapi.HeartbeatReq) (*logicapi.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(ctx, logicapi.MidType(req.Mid), req.Key, req.Server); err != nil {
		return &logicapi.HeartbeatReply{}, err
	}
	return &logicapi.HeartbeatReply{}, nil
}

// RenewOnline renew server online.
func (s *server) RenewOnline(ctx context.Context, req *logicapi.OnlineReq) (*logicapi.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(ctx, req.Server, req.RoomCount)
	if err != nil {
		return &logicapi.OnlineReply{}, err
	}
	return &logicapi.OnlineReply{AllRoomCount: allRoomCount}, nil
}

// Receive receive a message.
func (s *server) Receive(ctx context.Context, req *logicapi.ReceiveReq) (*logicapi.ReceiveReply, error) {
	if err := s.srv.Receive(ctx, req.DeviceId, logicapi.MidType(req.Mid), req.Platform, req.Proto); err != nil {
		return &logicapi.ReceiveReply{}, err
	}
	return &logicapi.ReceiveReply{}, nil
}

// nodes return nodes.
func (s *server) Nodes(ctx context.Context, req *logicapi.NodesReq) (*logicapi.NodesReply, error) {
	return s.srv.NodesWeighted(ctx, req.Platform, req.ClientIP), nil
}
