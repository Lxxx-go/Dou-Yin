package main

import (
	"fmt"
	commentservice "github.com/Lxxx-go/Dou-Yin/app/comment/cmd"
	favorservice "github.com/Lxxx-go/Dou-Yin/app/favorite/cmd"
	client "github.com/Lxxx-go/Dou-Yin/app/gateway/cmd"
	messageservice "github.com/Lxxx-go/Dou-Yin/app/message/cmd"
	relationservice "github.com/Lxxx-go/Dou-Yin/app/relation/cmd"
	userservice "github.com/Lxxx-go/Dou-Yin/app/user/cmd"
	"github.com/Lxxx-go/Dou-Yin/app/user/pkg/snowflake"
	"github.com/Lxxx-go/Dou-Yin/app/user/redis"
	videoservice "github.com/Lxxx-go/Dou-Yin/app/video/cmd"
	"github.com/Lxxx-go/Dou-Yin/conf"
	transfer "github.com/Lxxx-go/Dou-Yin/conf/confagent/log_transfer"
	"github.com/Lxxx-go/Dou-Yin/repo"
	"time"

	"go.uber.org/zap"
)

func main() {

	//1. 初始化配置文件
	if err := conf.Init(); err != nil {
		fmt.Printf("Config file initialization error,%#v", err)
		return
	}
	//2.初始化kafka消费者和ES
	go transfer.Init()

	//3. 初始化mysql数据库
	if err := repo.Init(); err != nil {
		fmt.Printf("mysql  init error,%#v", err)
		zap.L().Error("初始化mysql数据库失败！！！")
		return
	}

	//defer repo.Close()

	//4.初始化redis连接
	if err := redis.Init(); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	// 程序退出关闭数据库连接
	defer redis.Close()

	//5. snowflake初始化
	if err := snowflake.Init(conf.Config.StartTime, conf.Config.MachineID); err != nil {
		zap.L().Error("snowflake initialization error", zap.Error(err))
		return
	}
	//6.启动日志收集
	go func() {
		err := conf.Collectlog()
		if err != nil {
			zap.L().Error("log collect error ,", zap.Error(err))
		}
	}()

	//6.启动服务（后续可将所有的服务单独写到一个文件）
	go userservice.Start()
	time.Sleep(time.Second)
	go videoservice.Start()
	go favorservice.Start()
	go messageservice.Start()
	go commentservice.Start()
	go relationservice.Start()
	client.Start()
}
