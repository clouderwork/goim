package logic

import (
	"context"

	"github.com/Terry-Mao/goim/internal/logic/model"

	log "github.com/golang/glog"
)

// PushKeys push a message by keys.
func (l *Logic) PushKeys(c context.Context, op int32, seq int32, keys []string, msg []byte) (err error) {
	servers, err := l.dao.ServersByKeys(c, keys)
	if err != nil {
		return
	}
	pushKeys := make(map[string][]string)
	for i, key := range keys {
		server := servers[i]
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	for server := range pushKeys {
		if err = l.dao.PushMsg(c, op, seq, server, pushKeys[server], msg); err != nil {
			return
		}
	}
	return
}

// PushMids push a message by mid.
func (l *Logic) PushMidsWithoutKeys(c context.Context, op int32, seq int32, mids []int64, withoutKeys map[string]struct{}, msg []byte) (err error) {
	keyServers, _, err := l.dao.KeysByMids(c, mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key == "" || server == "" {
			log.Warningf("push key:%s server:%s is empty", key, server)
			continue
		}
		if _, ok := withoutKeys[key]; ok {
			continue
		}
		keys[server] = append(keys[server], key)
	}
	for server, keys := range keys {
		if err = l.dao.PushMsg(c, op, seq, server, keys, msg); err != nil {
			return
		}
	}
	return
}

// PushMids push a message by mid.
func (l *Logic) PushMids(c context.Context, op int32, seq int32, mids []int64, msg []byte) (err error) {
	keyServers, _, err := l.dao.KeysByMids(c, mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key == "" || server == "" {
			log.Warningf("push key:%s server:%s is empty", key, server)
			continue
		}
		keys[server] = append(keys[server], key)
	}
	for server, keys := range keys {
		if err = l.dao.PushMsg(c, op, seq, server, keys, msg); err != nil {
			return
		}
	}
	return
}

// PushRoom push a message by room.
func (l *Logic) PushRoom(c context.Context, op int32, seq int32, typ, room string, msg []byte) (err error) {
	return l.dao.BroadcastRoomMsg(c, op, seq, model.EncodeRoomKey(typ, room), msg)
}

// PushAll push a message to all.
func (l *Logic) PushAll(c context.Context, op, seq, speed int32, msg []byte) (err error) {
	return l.dao.BroadcastMsg(c, op, seq, speed, msg)
}
