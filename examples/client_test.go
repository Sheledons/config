package examples

import (
	"fmt"
	"github.com/Sheledons/config/sdk"
	"testing"
)

/**
  @author: baisongyuan
  @since: 2022/7/25
**/
func TestSDK(t *testing.T) {
	influxConf := sdk.NewUConfiguration("/xunlei/huiyuan/xluser/pod1/mysql",func(conf string) {
		fmt.Printf("xunlei mysql config update: %s\n", conf)
	})
	mysqlConf := sdk.NewUConfiguration("/wangxing/lian/xlppc/pod2/mysql",func(conf string) {
		fmt.Printf("wangxing mysql config update: %s\n", conf)
	})
	client := sdk.NewClient("测试client", ":18848", influxConf, mysqlConf)
	client.StartSubscribe()
	ch := make(chan int)
	<-ch
}
