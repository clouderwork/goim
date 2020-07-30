package grpc

import "strconv"

func (r *RoomsReply) XString() string {
	return strconv.FormatInt(int64(len(r.Rooms)), 10)
}
