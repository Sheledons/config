package sdk

import (
	"context"
	"github.com/Sheledons/config/config-sdk/pkg/log"
	"github.com/Sheledons/config/config-sdk/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

/**
  @author: baisongyuan
  @since: 2022/7/24
**/

type client struct {
	name        string
	conn        *grpc.ClientConn
	addr        string
	errChannel  chan error
	subscribers []*Subscriber
}

func NewClient(name, addr string, confs ...UConfiuration) *client {
	//defer func() {
	//	if r:=recover();r != nil {
	//		log.Sugar.Errorf("init client fail. %v",r)
	//	}
	//}()
	var subscribers []*Subscriber
	var sub *Subscriber
	for i, conf := range confs {
		// 10个一组
		if i%10 == 0 {
			sub = NewSubscriber()
			subscribers = append(subscribers, sub)
		}
		sub.addConfiguration(newConfiguration(conf.key, conf.callback))
	}
	return &client{
		name:        name,
		errChannel:  make(chan error, 512),
		addr:        addr,
		subscribers: subscribers,
	}
}

//TCP 连接由 clientConn.go 来维护，但是流并不能够自动被处理。
//流一旦断开，无论是由于 RPC 连接断开还是其他原因，都无法自动重新连接，
//并且需要在 RPC 连接备份后从服务器获取新流。

func (c *client) StartSubscribe() {
	var err error
	c.conn, err = grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Sugar.Errorf(c.addr)
		return
	}
	sc := protocol.NewStreamConfigServiceClient(c.conn)
	for _, sub := range c.subscribers {
		go c.process(sub, sc)
	}
}
func (c *client) process(sub *Subscriber, sc protocol.StreamConfigServiceClient) {
	go sub.startSubscribe(c.name, sc, c.errChannel)
	go func() { // 监控协程，每个subscriber只会有一个
		for {
			<-c.errChannel // 注意: 客户端和服务端交互过程中的业务异常不要抛到这里，比如握手验证失败之类的;只对GRPC产生的异常负责
			for !c.waitUntilReady() {
			} // 重新连接上层 Connection 等待其 READY
			// 连接恢复之后重置errChannel,通过传递引用避免线程安全问题。将client的errChannel下发到下层stream
			c.resetErrChannel()
			go sub.startSubscribe(c.name, sc, c.errChannel) // grpc 产生的错误，重连处理。
		}
	}()
}
func (c *client) waitUntilReady() bool { // connection 重新连接之后，流的重新建立才有意义
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) //define how long you want to wait for connection to be restored before giving up
	defer cancel()
	currentState := c.conn.GetState()
	stillConnecting := true
	for currentState != connectivity.Ready && stillConnecting {
		//will return true when state has changed from thisState, false if timeout
		//CONNECTING,TRANSIENT_FAILURE,READY return true
		//IDLE or timeout return false
		c.conn.Connect()
		c.conn.ResetConnectBackoff()
		stillConnecting = c.conn.WaitForStateChange(ctx, currentState)
		currentState = c.conn.GetState()
	}
	if !stillConnecting {
		log.Sugar.Error("Connection attempt has timed out.")
	}
	return stillConnecting
}
func (c *client) resetErrChannel() {
	close(c.errChannel)
	c.errChannel = make(chan error)
}
