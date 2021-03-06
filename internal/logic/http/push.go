package http

import (
	"context"
	"io/ioutil"

	logicapi "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/gin-gonic/gin"
)

func (s *Server) pushKeys(c *gin.Context) {
	var arg struct {
		Op   int32    `form:"operation"`
		Seq  int32    `form:"seq"`
		Ver  int32    `form:"ver"`
		Keys []string `form:"keys"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushKeys(context.TODO(), arg.Op, arg.Seq, arg.Ver, arg.Keys, msg); err != nil {
		result(c, nil, RequestErr)
		return
	}
	result(c, nil, OK)
}

func (s *Server) pushMidsWithoutKeys(c *gin.Context) {
	var arg struct {
		Op   int32              `form:"operation"`
		Seq  int32              `form:"seq"`
		Ver  int32              `form:"ver"`
		Mids []logicapi.MidType `form:"mids"`
		Keys []string           `form:"keys"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}

	withoutKeys := make(map[string]struct{}, len(arg.Keys))
	for _, key := range arg.Keys {
		withoutKeys[key] = struct{}{}
	}
	// PushMidsWithoutKeys(c context.Context, op int32, seq int32, mids []int64, withoutKeys map[string]struct{}, msg []byte) (err error)
	if err = s.logic.PushMidsWithoutKeys(context.TODO(), arg.Op, arg.Seq, arg.Ver, arg.Mids, withoutKeys, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

func (s *Server) pushMids(c *gin.Context) {
	var arg struct {
		Op   int32              `form:"operation"`
		Seq  int32              `form:"seq"`
		Ver  int32              `form:"ver"`
		Mids []logicapi.MidType `form:"mids"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushMids(context.TODO(), arg.Op, arg.Seq, arg.Ver, arg.Mids, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

func (s *Server) pushRoom(c *gin.Context) {
	var arg struct {
		Op   int32  `form:"operation" binding:"required"`
		Seq  int32  `form:"seq"`
		Ver  int32  `form:"ver"`
		Type string `form:"type" binding:"required"`
		Room string `form:"room" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushRoom(c, arg.Op, arg.Seq, arg.Ver, arg.Type, arg.Room, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

func (s *Server) pushAll(c *gin.Context) {
	var arg struct {
		Op    int32 `form:"operation" binding:"required"`
		Seq   int32 `form:"seq"`
		Ver   int32 `form:"ver"`
		Speed int32 `form:"speed"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushAll(c, arg.Op, arg.Seq, arg.Ver, arg.Speed, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}
