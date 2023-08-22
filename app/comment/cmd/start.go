package commentservice

import (
	"github.com/Lxxx-go/Dou-Yin/app/comment/service"
	"github.com/Lxxx-go/Dou-Yin/conf"
	"github.com/Lxxx-go/Dou-Yin/discovery"

	pb "github.com/Lxxx-go/Dou-Yin/protos/comment"

	"go.uber.org/zap"
)

func Start() {

	zap.L().Info("服务启动，开始记录日志")

	serviceRegister, grpcServer := discovery.InitRegister(conf.Config.CommentServiceName, conf.Config.CommentServiceUrl)
	defer serviceRegister.Close()
	defer grpcServer.Stop()
	pb.RegisterDouyinCommentServiceServer(grpcServer, &service.CommentSrv{}) // 绑定grpc
	discovery.GrpcListen(grpcServer, conf.Config.CommentServiceUrl)          // 开启监听

}
