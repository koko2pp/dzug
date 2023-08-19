package etcd

import (
	"context"
	"dzug/conf/tailfile"
	"dzug/models"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"time"
)

// etcd 相关操作

var (
	client *clientv3.Client
)

// Init 初始化  etcd
func Init(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v", err)
		return
	}
	return
}

// GetProjectConf 拉取日志收集配置项的函数
func GetProjectConf(key string) (config *models.ProjectConfig, err error) {
	var configlist []*models.ProjectConfig

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	resp, err := client.Get(ctx, key)
	if err != nil {
		//zap.L().Error("get conf from etcd by key:" + fmt.Sprintf("%s", key) + " failed ,err:%v" + fmt.Sprintf("%s", err))
		fmt.Printf("get conf from etcd by key:%s ,err:%v", key, err)
		return
	}
	if len(resp.Kvs) == 0 {
		//zap.L().Warn("get len:0 conf from etcd by key:%s" + fmt.Sprintf("%s", key))
		fmt.Printf("get len:0 conf from etcd by key:%s", key)
		return
	}

	ret := resp.Kvs[0] //取一个
	// ret.Value // json格式字符串
	fmt.Printf("%s", ret.Value)

	//将从etcd中去取出来的值ret.Value利用Unmarshal方法反序列化出来，存放在collectEntryList上
	err = json.Unmarshal(ret.Value, &configlist)
	if err != nil {
		//zap.L().Error("json unmarshal failed, err:", zap.Error(err))
		fmt.Printf("json unmarshal failed, err:%v", err)
		return nil, err
	}
	config = configlist[0]
	return
}

// GetProjectConf 拉取日志收集配置项的函数
func GetLogConf(key string) (logconflist []*models.LogConfig, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	resp, err := client.Get(ctx, key)
	if err != nil {
		//zap.L().Error("get conf from etcd by key:" + fmt.Sprintf("%s", key) + " failed ,err:%v" + fmt.Sprintf("%s", err))
		fmt.Printf("get conf from etcd by key:%s ,err:%v", key, err)
		return
	}
	if len(resp.Kvs) == 0 {
		//zap.L().Warn("get len:0 conf from etcd by key:%s" + fmt.Sprintf("%s", key))
		fmt.Printf("get len:0 conf from etcd by key:%s", key)
		return
	}

	ret := resp.Kvs[0] //取一个
	// ret.Value // json格式字符串
	fmt.Printf("%s", ret.Value)

	//将从etcd中去取出来的值ret.Value利用Unmarshal方法反序列化出来，存放在collectEntryList上
	err = json.Unmarshal(ret.Value, &logconflist)
	if err != nil {
		//zap.L().Error("json unmarshal failed, err:", zap.Error(err))
		fmt.Printf("json unmarshal failed, err:%v", err)
		return nil, err
	}
	return
}

// WatchConf 监控etcd中日志收集项配置变化的函数
func WatchConf(key string) {
	for {
		watchCh := client.Watch(context.Background(), key)
		for wresp := range watchCh {
			zap.L().Info("get new conf from etcd!")
			for _, evt := range wresp.Events {
				fmt.Printf("type:%s key:%s value:%s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
				var newConf []models.LogConfig

				if evt.Type == clientv3.EventTypeDelete {
					// 如果是 删除事件
					zap.L().Warn("FBI warning:etcd delete the key!!!")
					tailfile.SendNewConf(newConf) // 没有任何接收就是阻塞的
					continue
				}

				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					zap.L().Error("json unmarshal new conf failed, err:", zap.Error(err))
					continue
				}
				// 告诉tailfile这个模块应该启用新的配置了!
				tailfile.SendNewConf(newConf) // 没有人接收就是阻塞的
			}
		}
	}
}