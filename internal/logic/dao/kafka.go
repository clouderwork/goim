package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	comet "github.com/Terry-Mao/goim/api/comet/grpc"
	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/clouderwork/workchat/api/pbrequest"
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	sarama "gopkg.in/Shopify/sarama.v1"
)

// PushMsg push a message to databus.
func (d *Dao) PushMsg(c context.Context, op int32, seq int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Seq:       seq,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
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
func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, seq int32, room string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Seq:       seq,
		Room:      room,
		Msg:       msg,
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
func (d *Dao) BroadcastMsg(c context.Context, op, seq, speed int32, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Seq:       seq,
		Speed:     speed,
		Msg:       msg,
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
func (d *Dao) Dispatch(c context.Context, mid int64, data *comet.Proto) (err error) {

	isProto := true

	request := &pbrequest.Req{}
	if err = proto.Unmarshal(data.Body, request); err != nil {
		if err = json.Unmarshal(data.Body, request); err != nil {
			return
		} else {
			isProto = false
		}
	}

	if request.Seq != data.Seq {
		request.Ver = data.Ver
		request.Op = data.Op
		request.Seq = data.Seq
		if isProto {
			data.Body, err = proto.Marshal(request)
		} else {
			data.Body, err = json.Marshal(request)
		}
		if err != nil {
			return
		}
	}

	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(mid, 10)),
		Topic: fmt.Sprintf("%s-%s", d.c.Kafka.CallTopicPre, request.Module),
		Value: sarama.ByteEncoder(data.Body),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("Dispatch.send(Call:%v) error(%v)", data, err)
	}
	return
}
