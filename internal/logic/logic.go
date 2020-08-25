package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/Terry-Mao/goim/internal/logic/dao"
	"github.com/Terry-Mao/goim/internal/logic/model"
	"github.com/bilibili/discovery/naming"
	ucli "github.com/clouderwork/workchat/clientlib/user"
	log "github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5

	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

// Logic struct
type Logic struct {
	c   *conf.Config
	dis *naming.Discovery
	dao *dao.Dao
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
	// load balancer
	nodes        []*naming.Instance
	loadBalancer *LoadBalancer
	regions      map[string]string // province -> region

	userClient ucli.UserServiceClient
}

func newUserClient(c *conf.UserRPCClient) ucli.UserServiceClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	var (
		conn *grpc.ClientConn
		err  error
	)
	if c.Clitype == "smart" {
		conn, err = grpc.DialContext(ctx, fmt.Sprintf("discovery://default/%s", c.Appid),
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithInitialWindowSize(grpcInitialWindowSize),
				grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
				grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
				grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                grpcKeepAliveTime,
					Timeout:             grpcKeepAliveTimeout,
					PermitWithoutStream: true,
				}),
				grpc.WithBalancerName(roundrobin.Name),
			}...)
	} else {
		conn, err = grpc.Dial(c.Addr,
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithInitialWindowSize(grpcInitialWindowSize),
				grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
				grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
				grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                grpcKeepAliveTime,
					Timeout:             grpcKeepAliveTimeout,
					PermitWithoutStream: true,
				}),
			}...)
	}
	if err != nil {
		panic(err)
	}
	return ucli.NewGRPCUserServiceClient(conn)
}

// New init
func New(c *conf.Config) (l *Logic) {
	uc := newUserClient(c.UserRPCClient)
	l = &Logic{
		c:            c,
		dao:          dao.New(c),
		dis:          naming.New(c.Discovery),
		loadBalancer: NewLoadBalancer(),
		regions:      make(map[string]string),
		userClient:   uc,
	}
	l.initRegions()
	l.initNodes()
	_ = l.loadOnline()
	go l.onlineproc()
	return l
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	l.dao.Close()
}

func (l *Logic) initRegions() {
	for region, ps := range l.c.Regions {
		for _, province := range ps {
			l.regions[province] = region
		}
	}
}

func (l *Logic) initNodes() {
	res := l.dis.Build("goim.comet")
	event := res.Watch()
	select {
	case _, ok := <-event:
		if ok {
			l.newNodes(res)
		} else {
			panic("discovery watch failed")
		}
	case <-time.After(10 * time.Second):
		log.Error("discovery start timeout")
	}
	go func() {
		for {
			if _, ok := <-event; !ok {
				return
			}
			l.newNodes(res)
		}
	}()
}

func (l *Logic) newNodes(res naming.Resolver) {
	if zoneIns, ok := res.Fetch(); ok {
		var (
			totalConns int64
			totalIPs   int64
			allIns     []*naming.Instance
		)
		for _, zins := range zoneIns.Instances {
			for _, ins := range zins {
				if ins.Metadata == nil {
					log.Errorf("node instance metadata is empty(%+v)", ins)
					continue
				}
				offline, err := strconv.ParseBool(ins.Metadata[model.MetaOffline])
				if err != nil || offline {
					log.Warningf("strconv.ParseBool(offline:%t) error(%v)", offline, err)
					continue
				}
				conns, err := strconv.ParseInt(ins.Metadata[model.MetaConnCount], 10, 32)
				if err != nil {
					log.Errorf("strconv.ParseInt(conns:%d) error(%v)", conns, err)
					continue
				}
				ips, err := strconv.ParseInt(ins.Metadata[model.MetaIPCount], 10, 32)
				if err != nil {
					log.Errorf("strconv.ParseInt(ips:%d) error(%v)", ips, err)
					continue
				}
				totalConns += conns
				totalIPs += ips
				allIns = append(allIns, ins)
			}
		}
		l.totalConns = totalConns
		l.totalIPs = totalIPs
		l.nodes = allIns
		l.loadBalancer.Update(allIns)
	}
}

func (l *Logic) onlineproc() {
	for {
		time.Sleep(_onlineTick)
		if err := l.loadOnline(); err != nil {
			log.Errorf("onlineproc error(%v)", err)
		}
	}
}

func (l *Logic) loadOnline() (err error) {
	var (
		roomCount = make(map[string]int32)
	)
	for _, server := range l.nodes {
		var online *model.Online
		online, err = l.dao.ServerOnline(context.Background(), server.Hostname)
		if err != nil {
			return
		}
		if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
			_ = l.dao.DelServerOnline(context.Background(), server.Hostname)
			continue
		}
		for roomID, count := range online.RoomCount {
			roomCount[roomID] += count
		}
	}
	l.roomCount = roomCount
	return
}
