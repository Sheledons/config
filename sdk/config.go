package sdk

import (
	"context"
	"fmt"
	"github.com/Sheledons/config/config-sdk/pkg/log"
	"github.com/Sheledons/config/config-sdk/pkg/protocol"
	"github.com/spaolacci/murmur3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
	"time"
)

/**
  @author: baisongyuan
  @since: 2022/7/24
**/

type callback func(conf string)

type configuration struct {
	Key string
	callback
}
type UConfiuration struct {
	key      string
	callback
}

func NewUConfiguration(name string, callback2 callback) UConfiuration {
	return UConfiuration{
		key:      name,
		callback:  callback2,
	}
}

func newConfiguration(key string, callback callback) configuration {
	return configuration{
		Key:      key,
		callback: callback,
	}
}

type Subscriber struct {
	CallbackMap  map[uint32]callback
	CallbackLock map[uint32]*sync.Mutex
	// key:metaId value: /path
	ConfigMetaMap  map[uint32]string
	Configurations []configuration
	stream         protocol.StreamConfigService_ConfigChannelClient
	taskQueue      *TaskQueue
	Context        context.Context
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		CallbackMap:   make(map[uint32]callback),
		CallbackLock:  make(map[uint32]*sync.Mutex),
		ConfigMetaMap: make(map[uint32]string),
		taskQueue:     newTaskQueue(),
		Context:       context.Background(),
	}
}
func generateConfigMetaId(key string) uint32 {
	return murmur3.Sum32([]byte(key))
}
func (subs *Subscriber) getConfigNameList() []string {
	var confNameList []string
	for _, conf := range subs.Configurations {
		confNameList = append(confNameList, conf.Key)
	}
	return confNameList
}
func (subs *Subscriber) startSubscribe(clientName string, streamConfigServiceClient protocol.StreamConfigServiceClient, errChannel chan<- error) {
	var err error
	var streamPacket *protocol.StreamPacket
	defer subs.destroy()
	// 在TCP连接中创建一个双向流
	subs.Context = context.Background()
	subs.stream, err = streamConfigServiceClient.ConfigChannel(subs.Context)
	if err != nil {
		log.Sugar.Errorf("rpc fail. %s\n", err)
		goto errHere
	}
	if err = handshake(subs, clientName); err != nil {
		log.Sugar.Errorf("handshake fail. %s", err)
		goto errHere
	}
	if err = pullConfig(subs); err != nil {
		goto errHere
	}
	// 周期任务之全量拉去
	subs.addTickerTask("full pull config from config center", 5*time.Minute, func() {
		_ = pullConfig(subs)
	})
	for {
		streamPacket, err = subs.stream.Recv()
		if err == io.EOF {
			// TODO: 配置中心主动调用断开连接，后面考虑怎么处理
			log.Sugar.Warnf("rpc stream EOF.\n")
			return
		}
		if err != nil {
			log.Sugar.Errorf("rpc stream err: %s\n", err)
			goto errHere
		}
		log.Sugar.Info("receive pack. %v", streamPacket)
		if handle, ok := HandleStrategy[streamPacket.Code]; ok {
			go handle(streamPacket, subs, errChannel)
		} else {
			log.Sugar.Warnf("unkown packet type. %v", streamPacket)
		}
	}
	// 业务异常不要写人 errChannel ！！！

errHere:
	fromError, ok := status.FromError(err)
	if ok && fromError.Code() != codes.Unknown {
		errChannel <- err
	} else {
		//业务上面的错误
		fmt.Println(err)
	}
}
func (subs *Subscriber) addTickerTask(name string, period time.Duration, task func()) {
	subs.taskQueue.addTask(name, period, task)
}
func (subs *Subscriber) destroy() {
	// 关闭旧的流
	if subs.stream != nil {
		_ = subs.stream.CloseSend()
		subs.Context.Done()
	}
	subs.taskQueue.stopAll()
}
func (subs *Subscriber) addConfiguration(conf configuration) {
	subs.Configurations = append(subs.Configurations, conf)
	key := generateConfigMetaId(conf.Key)
	subs.CallbackMap[key] = conf.callback
	subs.CallbackLock[key] = new(sync.Mutex)
	subs.ConfigMetaMap[key] = conf.Key
}
