package main

import (
	"cosgo/app"
	"cosgo/cosnet"
	"cosgo/logger"
	"github.com/spf13/pflag"
)

func init() {
	pflag.String("tcp", "tcp://0.0.0.0:3100", "tcp address")
}
func NewTcpModule(id string) *tcpModule {
	return &tcpModule{
		DefModule: app.DefModule{Id: id},
	}
}

type tcpModule struct {
	app.DefModule
	srv cosnet.Server
}

func (m *tcpModule) Init() (err error) {
	addr := app.Config.GetString("tcp")
	m.srv, err = cosnet.NewServer(addr, &TcpHandler{})
	return
}

func (m *tcpModule) Start() error {
	m.srv.Start()
	return nil
}
func (m *tcpModule) Close() error {
	return m.srv.Close()
}

type TcpHandler struct {
	cosnet.HandlerDefault
}

func (this *TcpHandler) OnMessage(sock cosnet.Socket, msg *cosnet.Message) bool {
	logger.Debug("OnMessage:%+v", msg)
	return true
}
