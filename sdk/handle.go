package sdk

import (
	"fmt"
	"github.com/Sheledons/config/config-sdk/pkg/log"
	"github.com/Sheledons/config/config-sdk/pkg/packet"
	"github.com/Sheledons/config/config-sdk/pkg/protocol"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

/**
  @author: baisongyuan
  @since: 2022/7/26
**/

type packetHandle func(*protocol.StreamPacket, *Subscriber, chan<- error)

var HandleStrategy = make(map[int32]packetHandle)

func HandleConfigChangePacket(p *protocol.StreamPacket, subscriber *Subscriber, errChannel chan<- error) {
	// 加入md5比对
	proConf := protocol.Configuration{}
	_ = p.Body.UnmarshalTo(&proConf)
	// check md5
	if configurationCache.CheckIsSame(proConf.MetaId, proConf.Md5) {
		log.Sugar.Infof("configuration same , not call callback function. %v",proConf.Name)
		return
	}
	wrapper := wrapperspb.UInt32(proConf.MetaId)
	anyB, _ := anypb.New(wrapper)
	if err := subscriber.stream.Send(&protocol.StreamPacket{
		Code: packet.RequestConfigPacket,
		Body: anyB,
	}); err != nil {
		log.Sugar.Errorf("config:[ %s ] request fail. ", subscriber.ConfigMetaMap[proConf.MetaId])
		errChannel <- err
	}
}
func HandleResponseConfig(p *protocol.StreamPacket, subscriber *Subscriber, errChannel chan<- error) {
	conf := protocol.Configuration{}
	if err := p.Body.UnmarshalTo(&conf); err != nil {
		log.Sugar.Errorf("response config parse fail. %s", err)
		return
	}
	if configurationCache.CheckIsSame(conf.MetaId, conf.Md5) {
		fmt.Printf("config [%d] not change \n",conf.MetaId)
		return
	}
	if _, ok := subscriber.CallbackMap[conf.MetaId]; ok {
		subscriber.CallbackLock[conf.MetaId].Lock()
		defer subscriber.CallbackLock[conf.MetaId].Unlock()
		subscriber.CallbackMap[conf.MetaId](conf.Content)
		configurationCache.AndOrUpdate(configurationModels{
			MetaId:  conf.MetaId,
			Content: conf.Content,
			md5:     conf.Md5,
			Name:    conf.Name,
		})
	}
}

func init() {
	HandleStrategy[packet.ConfigChangePacket] = HandleConfigChangePacket
	HandleStrategy[packet.ResponseConfig] = HandleResponseConfig
}
