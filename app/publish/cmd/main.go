package main

import (
	"fmt"
	"github.com/Lxxx-go/Dou-Yin/app/publish/infra/redis"
	"github.com/Lxxx-go/Dou-Yin/app/publish/pkg/oss"
	"github.com/Lxxx-go/Dou-Yin/app/publish/service"
	"github.com/Lxxx-go/Dou-Yin/conf"
	"github.com/Lxxx-go/Dou-Yin/discovery"
	"github.com/Lxxx-go/Dou-Yin/logger"
	pb "github.com/Lxxx-go/Dou-Yin/protos/publish"
	"github.com/Lxxx-go/Dou-Yin/repo"
	"go.uber.org/zap"
)

func main() {
	if err := conf.Init(); err != nil {
		fmt.Printf("Config file initialization error,%#v", err)
		return
	}

	if err := logger.Init(conf.Config.LogConfig); err != nil {
		fmt.Printf("log file initialization error,%#v", err)
		return
	}

	defer zap.L().Sync()
	zap.L().Info("服务启动，开始记录日志")

	// 初始化redis
	err := redis.Init()
	if err != nil {
		return
	}
	oss.Init()
	_ = repo.Init()
	serviceRegister, grpcServer := discovery.InitRegister(conf.Config.PublishServiceName, conf.Config.PublishServiceUrl)
	defer serviceRegister.Close()
	defer grpcServer.Stop()
	pb.RegisterPublishServiceServer(grpcServer, &service.VideoServer{})
	discovery.GrpcListen(grpcServer, conf.Config.PublishServiceUrl)
}
