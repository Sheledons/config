package log

import "go.uber.org/zap"

/**
  @author: baisongyuan
  @since: 2022/7/25
**/

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {
	Logger, _ = zap.NewProduction()
	Sugar = Logger.Sugar()
}
