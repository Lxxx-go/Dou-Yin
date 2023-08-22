package videoservice

import (
	"github.com/Lxxx-go/Dou-Yin/app/video/service"
	"github.com/Lxxx-go/Dou-Yin/conf"
	"github.com/Lxxx-go/Dou-Yin/discovery"
	"github.com/Lxxx-go/Dou-Yin/protos/video"
)

func Start() {

	// 传入注册的服务名和注册的服务地址进行注册
	serviceRegister, grpcServer := discovery.InitRegister(conf.Config.VideoServiceName, conf.Config.VideoServiceUrl)
	defer serviceRegister.Close()
	defer grpcServer.Stop()
	video.RegisterVideoServiceServer(grpcServer, &service.VideoService{}) // 绑定grpc
	discovery.GrpcListen(grpcServer, conf.Config.VideoServiceUrl)         // 开启监听

}
