package cosnet

import (
	"context"
	"github.com/hwcer/cosgo/cosnet/message"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/storage"
	"github.com/hwcer/cosgo/utils"
	"sync/atomic"
)

//各种服务器(TCP,UDP,WS)也使用该接口
type Socket interface {
	Id() uint64
	Set(interface{})  //设置USER DATA
	Get() interface{} //获取USER DATA
	Close() bool
	Write(m *message.Message) bool
	IsProxy() bool
	Heartbeat()
	KeepAlive()
	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)
}

func NewNetSocket(s Server) *NetSocket {
	sock := &NetSocket{
		cwrite: make(chan *message.Message, Config.WriteChanSize),
		server: s,
	}
	sock.ctx, sock.cancel = context.WithCancel(s.Context())
	return sock
}

type NetSocket struct {
	ctx            context.Context       //context
	cancel         context.CancelFunc    //cancel
	status         int32                 //0:正常，1:等待断线重连，2:已经关闭
	server         Server                //server
	cwrite         chan *message.Message //写入通道
	heartbeat      int                   //heartbeat >=timeout 时被标记为超时
	realRemoteAddr string                //当使用代理是，需要特殊设置客户端真实IP
	*storage.ArrayDatasetDefault
}

//关闭
func (s *NetSocket) Close() bool {
	var newStatus int32
	if Config.ReconnectTime > 0 {
		newStatus = 1
	} else {
		newStatus = 2
	}
	if !atomic.CompareAndSwapInt32(&s.status, 0, newStatus) {
		return false
	}
	s.cancel()
	logger.Debug("Socket Finish Id:%d", s.Id())
	return true
}

func (s *NetSocket) Stopped() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
}

//Heartbeat 每一次Heartbeat() heartbeat计数加1
func (s *NetSocket) Heartbeat() {
	s.heartbeat += 1
	if s.heartbeat >= Config.SocketTimeout {
		s.cancel()
	}
}

//KeepAlive 任何行为都清空heartbeat
func (s *NetSocket) KeepAlive() {
	s.heartbeat = 0
}

func (s *NetSocket) LocalAddr() string {
	return ""
}
func (s *NetSocket) RemoteAddr() string {
	return ""
}

func (s *NetSocket) SetRealRemoteAddr(addr string) {
	s.realRemoteAddr = addr
}

func (s *NetSocket) Write(m *message.Message) (re bool) {
	if m == nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			re = false
		}
	}()
	if Config.AutoCompressSize > 0 && m.Head != nil && m.Head.Size >= Config.AutoCompressSize && !m.Head.Flags.Has(message.FlagCompress) {
		m.Head.Flags.Set(message.FlagCompress)
		m.Data = utils.GZipCompress(m.Data)
		m.Head.Size = int32(len(m.Data))
	}
	select {
	case s.cwrite <- m:
	default:
		logger.Warn("socket write channel full id:%v", s.Id())
		s.cancel() //通道已满，直接关闭
	}
	return true
}

func (s *NetSocket) processMsg(sock Socket, msg *message.Message) {
	s.KeepAlive()
	if msg.Head != nil && msg.Head.Flags.Has(message.FlagCompress) && msg.Data != nil {
		data, err := utils.GZipUnCompress(msg.Data)
		if err != nil {
			s.cancel()
			logger.Error("uncompress failed socket:%v err:%v", sock.Id(), err)
			return
		}
		msg.Data = data
		msg.Head.Flags.Remove(message.FlagCompress)
		msg.Head.Size = int32(len(msg.Data))
	}
	handler := s.server.Handler()
	handler.Message(sock, msg)
}
