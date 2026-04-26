package uuid

import "fmt"

var uuidWorker *UuidWorker

func init() {
	uuidWorker, _ = NewUuidWorker(0) // 默认 workerID=0
}

// InitWorker 初始化 workerID (0-1023)，由 HttpServerConfig.WorkerID 配置
func InitWorker(workerID int64) error {
	var err error
	uuidWorker, err = NewUuidWorker(workerID)
	return err
}

func GetUuidString() string {
	return fmt.Sprintf("%d", uuidWorker.Get())
}

// GenSessionId 生成全局唯一的 sessionId
func GenSessionId() string {
	return fmt.Sprintf("%d", uuidWorker.Get())
}

func GetUuidInt64() int64 {
	return uuidWorker.Get()
}
