package sdk

import (
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"strings"
	"xconfig-sdk-go/pkg/log"
	"xconfig-sdk-go/pkg/packet"
	"xconfig-sdk-go/pkg/protocol"
)

/**
  @author: baisongyuan
  @since: 2022/7/26
**/

func handshake(subs *Subscriber, clientName string) error {
	handshakeBody, _ := anypb.New(&protocol.ClientHandShake{
		ClientName:    clientName,
		SubConfigList: subs.getConfigNameList(),
	})
	if err := subs.stream.Send(&protocol.StreamPacket{
		Code: packet.Handshake,
		Id:   generateId(),
		Body: handshakeBody,
	}); err != nil {
		return err
	}
	streamPacket, err := subs.stream.Recv() // 握手的响应
	log.Sugar.Debug("receive pack. %v", streamPacket)
	if err != nil {
		return err
	}
	if streamPacket.Code != packet.HandshakeSuccess {
		return fmt.Errorf("handshake fail. %s", err.Error())
	}
	return nil
}

func pullConfig(subs *Subscriber) error {
	for metaId, conf := range subs.ConfigMetaMap {
		wrapperBody := wrapperspb.UInt32(metaId)
		anyBody, _ := anypb.New(wrapperBody)
		if err := subs.stream.Send(&protocol.StreamPacket{
			Code: packet.RequestConfigPacket,
			Body: anyBody,
		}); err != nil {
			log.Sugar.Errorf("send pull config [%s] packet fail. %s", conf, err)
			return err
		}
	}
	return nil
}
func generateId() string {
	newUUID, _ := uuid.NewUUID()
	return strings.Replace(newUUID.String(), "-", "", 1)[0:10]
}
