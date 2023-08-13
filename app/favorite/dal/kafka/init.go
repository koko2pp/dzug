package kafka

import (
	"github.com/IBM/sarama"
)

var (
	KafkaConsumer sarama.Consumer              // 消费者
	KafkaProducer sarama.SyncProducer          // 生产者
	KafkaAddr     = []string{"127.0.0.1:9092"} // kafka地址，之后写进config.yaml里去
)

func InitKafka() {
	var err error                                           // 成功交付的消息将在success channel返回
	KafkaConsumer, err = sarama.NewConsumer(KafkaAddr, nil) // 启动消费者，此时消费者开始消费
	if err != nil {
		panic("kafka错误：" + err.Error())
	}

	config := sarama.NewConfig()                            // 配置生成这
	config.Producer.RequiredAcks = sarama.WaitForAll        // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewHashPartitioner // 生产者分区用的hash分配，我暂时只用一个分区
	config.Producer.Return.Successes = true
	KafkaProducer, err = sarama.NewSyncProducer(KafkaAddr, config) // 初始化
	if err != nil {
		panic(err)
		return
	}
}
