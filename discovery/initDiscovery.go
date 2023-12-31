package discovery

import (
	"errors"
	"github.com/Lxxx-go/Dou-Yin/conf"
	"github.com/Lxxx-go/Dou-Yin/protos/comment"
	"github.com/Lxxx-go/Dou-Yin/protos/favorite"
	"github.com/Lxxx-go/Dou-Yin/protos/message"
	"github.com/Lxxx-go/Dou-Yin/protos/publish"
	"github.com/Lxxx-go/Dou-Yin/protos/relation"
	"github.com/Lxxx-go/Dou-Yin/protos/user"
	"github.com/Lxxx-go/Dou-Yin/protos/video"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	SerDiscovery   serviceDiscovery
	UserClient     user.ServiceClient
	VideoClient    video.VideoServiceClient
	MessageClient  message.DouyinMessageServiceClient
	FavoriteClient favorite.DouyinFavoriteActionServiceClient
	RelationClient relation.DouyinRelationActionServiceClient
	CommentClient  comment.DouyinCommentServiceClient
	PublishClient  publish.PublishServiceClient
)

// InitDiscovery 初始化一个服务发现程序
func InitDiscovery() {
	endpoints := conf.BasicConf.EtcdAddr                  // etcd地址
	SerDiscovery = serviceDiscovery{EtcdAddrs: endpoints} // 放入etcd地址
	err := SerDiscovery.newServiceDiscovery()             // 实例化
	if err != nil {
		zap.L().Fatal("启动服务发现失败: " + err.Error())
		return
	}
}

// LoadClient 加载etcd客户端调用实例，每一次客户端调用一个方法都会调用这个方法
// 先去etcd中拿去现在的链接，再去通过grpc进行远程调用
func LoadClient(serviceName string, client any) error {
	conn, err := connectService(serviceName) // 找到grpc连接链接
	if err != nil {
		zap.L().Error("grpc连接服务: " + serviceName + "失败, error: " + err.Error())
		return err
	}

	switch c := client.(type) {
	case *user.ServiceClient:
		*c = user.NewServiceClient(conn)
	case *video.VideoServiceClient:
		*c = video.NewVideoServiceClient(conn)
	case *favorite.DouyinFavoriteActionServiceClient:
		*c = favorite.NewDouyinFavoriteActionServiceClient(conn)
	case *relation.DouyinRelationActionServiceClient:
		*c = relation.NewDouyinRelationActionServiceClient(conn)
	case *comment.DouyinCommentServiceClient:
		*c = comment.NewDouyinCommentServiceClient(conn)
	case *message.DouyinMessageServiceClient:
		*c = message.NewDouyinMessageServiceClient(conn)
	case *publish.PublishServiceClient:
		*c = publish.NewPublishServiceClient(conn)
	default:
		err = errors.New("没有该类型的服务")
		zap.L().Error(err.Error())
	}
	return err
}

// connectService 通过服务名字找到对应的链接
// 比如，传入user，会找到etcd上存储的user的链接
func connectService(serviceName string) (conn *grpc.ClientConn, err error) {
	err = SerDiscovery.watchService("") // ！！！监视所有的服务
	if err != nil {
		zap.L().Error("未找到服务地址：" + err.Error())
		return nil, err
	}
	addr, err := SerDiscovery.getServiceByKey(serviceName)
	if err != nil {
		zap.L().Error("未找到服务地址：" + err.Error())
		return nil, err
	}
	conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return
}
