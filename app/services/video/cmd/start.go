package videoservice

import (
	"context"
	"dzug/app/services/video/service"
	"dzug/conf"
	"dzug/discovery"
	"dzug/models"
	"dzug/protos/video"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var VideoClient *clientv3.Client
var VideoBaseConf = new(models.BasicConfig)

func Start() (err error) {

	//1.初始化viper
	ymlPath := "/app/services/video/conf/config.yml"
	if err = conf.ViperInit(VideoBaseConf, ymlPath); err != nil {
		fmt.Printf("viper 初始化失败..., err:%v\n", err)
	}
	//2.连接etcd
	VideoClient, err = clientv3.New(clientv3.Config{
		Endpoints:   VideoBaseConf.EtcdAddr,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v", err)
		return
	}
	//3. 判断video配置是否已经存到etcd
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	resp, err := VideoClient.Get(ctx, VideoBaseConf.Name)
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	//如果video配置没有存到etcd
	if len(resp.Kvs) == 0 {
		//从yml文件中读取配置，存到etcd中
		if err := conf.ViperInit(conf.VideoConf, ymlPath); err != nil {
			fmt.Printf("viper 初始化失败..., err:%v\n", err)
		}
		err = conf.PutConfigToEtcd(VideoBaseConf.Name, conf.VideoConf)
		if err != nil {
			fmt.Println("user配置存到etcd过程中出错：" + err.Error())
			return
		}
	} else { //如果已经存到etcd上
		err = json.Unmarshal(resp.Kvs[0].Value, &conf.VideoConf)
	}
	//4.启动配置监控
	go WatchVideoConf(VideoBaseConf.Name)

	// 传入注册的服务名和注册的服务地址进行注册
	serviceRegister, grpcServer := discovery.InitRegister(conf.VideoConf.ServiceName, conf.VideoConf.Url)
	defer serviceRegister.Close()
	defer grpcServer.Stop()
	video.RegisterVideoServiceServer(grpcServer, &service.VideoService{}) // 绑定grpc
	discovery.GrpcListen(grpcServer, conf.VideoConf.Url)                  // 开启监听
	return
}

// WatchVideoConf 监控etcd中video服务配置变化
func WatchVideoConf(key string) {
	for {
		watchCh := VideoClient.Watch(context.Background(), key)
		for wresp := range watchCh {
			fmt.Println("get new conf from etcd!!!")
			for _, evt := range wresp.Events {
				//fmt.Printf("type:%s key:%s value:%s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
				err := json.Unmarshal(evt.Kv.Value, &conf.VideoConf)
				if err != nil {
					fmt.Println("json unmarshal new conf failed, err: ", err)
					continue
				}
			}
		}
	}
}
