package storage

import "context"

var Options = struct {
	MaxAge    int64 //有效期(S)
	MapSize   int32
	Heartbeat int32 //心跳(S)
}{
	MaxAge:    3600,
	MapSize:   1024,
	Heartbeat: 10,
}

type Dataset interface {
	Id() string
	Set(key string, val interface{})
	Get(key string) (interface{}, bool)
	Lock() bool
	Reset(bool) //自动续约,参数为true时解除锁定
	Expire() int64
}

type Storage interface {
	Get(string) (Dataset, bool)
	Start(ctx context.Context)
	Close()
	Create(map[string]interface{}) Dataset
	Delete(string) bool
}
