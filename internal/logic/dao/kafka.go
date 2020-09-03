package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	comet "github.com/Terry-Mao/goim/api/comet/grpc"
	logicapi "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/clouderwork/workchat/api/pbrequest"
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	sarama "gopkg.in/Shopify/sarama.v1"
)

// PushMsg push a message to databus.
func (d *Dao) PushMsg(c context.Context, op int32, seq int32, ver int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &logicapi.PushMsg{
		Type:      logicapi.PushMsg_PUSH,
		Operation: op,
		Seq:       seq,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
		Ver:       ver,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(keys[0]),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(push pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// BroadcastRoomMsg push a message to databus.
func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, seq int32, ver int32, room string, msg []byte) (err error) {
	pushMsg := &logicapi.PushMsg{
		Type:      logicapi.PushMsg_ROOM,
		Operation: op,
		Seq:       seq,
		Room:      room,
		Msg:       msg,
		Ver:       ver,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(room),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// BroadcastMsg push a message to databus.
func (d *Dao) BroadcastMsg(c context.Context, op, seq int32, ver int32, speed int32, msg []byte) (err error) {
	pushMsg := &logicapi.PushMsg{
		Type:      logicapi.PushMsg_BROADCAST,
		Operation: op,
		Seq:       seq,
		Speed:     speed,
		Msg:       msg,
		Ver:       ver,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(int64(op), 10)),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(broadcast pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// BroadcastMsg push a message to databus.
func (d *Dao) Dispatch(c context.Context, deviceID string, mid logicapi.MidType, platform string, data *comet.Proto) (err error) {

	isProtobuf := (data.Ver%2 == 0) // 借用ver字段求模，来区分数据序列化协议，目前只有protobuf和json，对2求模，是0和1；如果业务场景有n种序列化协议，那么对n求模；或者改造comet.Proto，加字段支持

	request := &pbrequest.Req{}
	if isProtobuf {
		err = proto.Unmarshal(data.Body, request)
	} else {
		err = json.Unmarshal(data.Body, request)
	}

	if err != nil {
		return
	}

	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(fmt.Sprintf("%s_%s_%s_%d_%d", deviceID, string(mid), platform, data.Ver, data.Seq)),
		Topic: fmt.Sprintf("%s-%s", d.c.Kafka.CallTopicPre, request.Module),
		Value: sarama.ByteEncoder(data.Body),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("Dispatch.send(Call:%v) error(%v)", data, err)
	}
	return
}
