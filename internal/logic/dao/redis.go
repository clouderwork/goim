package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	logicapi "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/Terry-Mao/goim/internal/logic/model"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"

	"github.com/zhenjl/cityhash"
)

const (
	_prefixMidServer    = "%s:mid_%s" // mid -> key:server
	_prefixKeyServer    = "%s:key_%s" // key -> server
	_prefixServerOnline = "%s:ol_%s"  // server -> online
)

func keyMidServer(namespace string, mid logicapi.MidType) string {
	return fmt.Sprintf(_prefixMidServer, namespace, string(mid))
}

func keyKeyServer(namespace string, key string) string {
	return fmt.Sprintf(_prefixKeyServer, namespace, key)
}

func keyServerOnline(namespace string, key string) string {
	return fmt.Sprintf(_prefixServerOnline, namespace, key)
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis.Get()
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

// AddMapping add a mapping.
// Mapping:
//	mid -> key_server
//	key -> server
func (d *Dao) AddMapping(c context.Context, mid logicapi.MidType, key, server string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	var n = 2
	if mid.IsNotZero() {
		if err = conn.Send("HSET", keyMidServer(d.redisPre, mid), key, server); err != nil {
			log.Errorf("conn.Send(HSET %s,%s,%s) error(%v)", mid, server, key, err)
			return
		}
		if err = conn.Send("EXPIRE", keyMidServer(d.redisPre, mid), d.redisExpire); err != nil {
			log.Errorf("conn.Send(EXPIRE %s,%s,%s) error(%v)", mid, key, server, err)
			return
		}
		n += 2
	}
	if err = conn.Send("SET", keyKeyServer(d.redisPre, key), server); err != nil {
		log.Errorf("conn.Send(HSET %s,%s,%s) error(%v)", mid, server, key, err)
		return
	}
	if err = conn.Send("EXPIRE", keyKeyServer(d.redisPre, key), d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < n; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// ExpireMapping expire a mapping.
func (d *Dao) ExpireMapping(c context.Context, mid logicapi.MidType, key string) (has bool, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	var n = 1
	if mid.IsNotZero() {
		if err = conn.Send("EXPIRE", keyMidServer(d.redisPre, mid), d.redisExpire); err != nil {
			log.Errorf("conn.Send(EXPIRE %s,%s) error(%v)", mid, key, err)
			return
		}
		n++
	}
	if err = conn.Send("EXPIRE", keyKeyServer(d.redisPre, key), d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s,%s) error(%v)", mid, key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < n; i++ {
		if has, err = redis.Bool(conn.Receive()); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DelMapping del a mapping.
func (d *Dao) DelMapping(c context.Context, mid logicapi.MidType, key, server string) (has bool, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	n := 1
	if mid.IsNotZero() {
		if err = conn.Send("HDEL", keyMidServer(d.redisPre, mid), key); err != nil {
			log.Errorf("conn.Send(HDEL %s,%s,%s) error(%v)", mid, key, server, err)
			return
		}
		n++
	}
	if err = conn.Send("DEL", keyKeyServer(d.redisPre, key)); err != nil {
		log.Errorf("conn.Send(HDEL %s,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < n; i++ {
		if has, err = redis.Bool(conn.Receive()); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// ServersByKeys get a server by key.
func (d *Dao) ServersByKeys(c context.Context, keys []string) (res []string, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	var args []interface{}
	for _, key := range keys {
		args = append(args, keyKeyServer(d.redisPre, key))
	}
	if res, err = redis.Strings(conn.Do("MGET", args...)); err != nil {
		log.Errorf("conn.Do(MGET %v) error(%v)", args, err)
	}
	return
}

// KeysByMids get a key server by mid.
func (d *Dao) KeysByMids(c context.Context, mids []logicapi.MidType) (ress map[string]string, olMids []logicapi.MidType, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	ress = make(map[string]string)
	for _, mid := range mids {
		if err = conn.Send("HGETALL", keyMidServer(d.redisPre, mid)); err != nil {
			log.Errorf("conn.Do(HGETALL %s) error(%v)", mid, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for idx := 0; idx < len(mids); idx++ {
		var (
			res map[string]string
		)
		if res, err = redis.StringMap(conn.Receive()); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
		if len(res) > 0 {
			olMids = append(olMids, mids[idx])
		}
		for k, v := range res {
			ress[k] = v
		}
	}
	return
}

// AddServerOnline add a server online.
func (d *Dao) AddServerOnline(c context.Context, server string, online *model.Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(d.redisPre, server)
	for hashKey, value := range roomsMap {
		err = d.addServerOnline(c, key, strconv.FormatInt(int64(hashKey), 10), &model.Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

func (d *Dao) addServerOnline(c context.Context, key string, hashKey string, online *model.Online) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, _ := json.Marshal(online)
	if err = conn.Send("HSET", key, hashKey, b); err != nil {
		log.Errorf("conn.Send(SET %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// ServerOnline get a server online.
func (d *Dao) ServerOnline(c context.Context, server string) (online *model.Online, err error) {
	online = &model.Online{RoomCount: map[string]int32{}}
	key := keyServerOnline(d.redisPre, server)
	for i := 0; i < 64; i++ {
		ol, err := d.serverOnline(c, key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return
}

func (d *Dao) serverOnline(c context.Context, key string, hashKey string) (online *model.Online, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, err := redis.Bytes(conn.Do("HGET", key, hashKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorf("conn.Do(HGET %s %s) error(%v)", key, hashKey, err)
		}
		return
	}
	online = new(model.Online)
	if err = json.Unmarshal(b, online); err != nil {
		log.Errorf("serverOnline json.Unmarshal(%s) error(%v)", b, err)
		return
	}
	return
}

// DelServerOnline del a server online.
func (d *Dao) DelServerOnline(c context.Context, server string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	key := keyServerOnline(d.redisPre, server)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorf("conn.Do(DEL %s) error(%v)", key, err)
	}
	return
}
