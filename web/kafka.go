package web

import (
	"IUS/conf"
	"encoding/json"

	//"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Shopify/sarama"
	//"golang.org/x/time/rate"
)

type KafkaAccess struct {
	Size  int64
	Path  string
	Chan  chan requests
	First int64
	Last  int64 // 上次发送信息
	Count int   // 数据总量
}

var (
	Address []string
	Topic   string
	Num     int
	Thread  int
)

func (access *KafkaAccess) Send(data []*sarama.ProducerMessage, client sarama.SyncProducer) {

	//发送消息
	err := client.SendMessages(data)
	if err != nil {
		log.Printf("send message failed: %v", err)
		return
	}
	log.Printf("Send %d data to kafka in %d", len(data), time.Now().Unix())
}

// 读取channel并发送
func (access *KafkaAccess) RChan(id int) {

	log.Printf("id is : %d\n", id)
	//kafka初始化配置
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	//新增lz4压缩方式
	config.Producer.Compression = sarama.CompressionGZIP
	//生产者
	client, err := sarama.NewSyncProducer(Address, config)
	if err != nil {
		log.Printf("producer close,err: %v", err)
		return
	}

	defer client.Close()
	s10 := time.NewTicker(20 * time.Second)
	defer func() {
		s10.Stop()
	}()

	// 可从配置文件里读
	count := 0
	access.Last = time.Now().Unix()
	data := make([]*sarama.ProducerMessage, Num)

	//l := rate.NewLimiter(rate.Limit(common.Conf.Limit/common.Conf.Thread), 1000)
	//c, _ := context.WithCancel(context.TODO())
	for {
		//l.Wait(c)
		select {
		case raw := <-access.Chan:
			jstr, err := json.Marshal(raw)
			//log.Printf("recv jstr : %v", jstr)
			if err == nil {
				msg := &sarama.ProducerMessage{}
				msg.Topic = Topic
				msg.Value = sarama.StringEncoder(fmt.Sprintf("%s", jstr))
				data[count] = msg
				count += 1
				if count == Num {
					access.Count += count
					count = 0
					access.Send(data, client)
					data = make([]*sarama.ProducerMessage, Num)
				}
			} else {
				log.Printf("json marshal error : %v", err)
			}
		case <-s10.C:
			if count != 0 {
				tmp := count
				count = 0
				access.Send(data[:tmp], client)
				access.Count += tmp
				data = make([]*sarama.ProducerMessage, Num)
				access.Last = time.Now().Unix()
			}
		}
	}
}

// 判断topic是否存在
func (access *KafkaAccess) Topic() error {

	var err error
	var cli sarama.Client
	var topics []string

	cli, err = sarama.NewClient(Address, sarama.NewConfig())
	defer func() {
		err = cli.Close()
		if err != nil {
			log.Printf("close kafka client error, %v", err)
		}
	}()
	if err != nil {
		log.Printf("create new client err : %v", err)
		return err
	}
	topics, err = cli.Topics()
	if err != nil {
		log.Printf("get topics err : %v", err)
		return err
	}
	for _, t := range topics {
		if t == Topic {
			return nil
		}
	}
	return errors.New("no topic : " + Topic)
}

// 启动线程
func (access *KafkaAccess) Thread() {

	Thread = conf.Cfg.Kafka.Thread

	log.Printf("%d new thread start.", Thread)
	for i := 0; i < Thread; i++ {
		go access.RChan(i)
	}
}

// 初始化
func (access *KafkaAccess) Start() {

	//s5 := time.NewTicker(5 * time.Second)
	//defer func() {
	//s5.Stop()
	//}()

	//for {
	//select {
	//case <-s5.C:
	Address = []string{conf.Cfg.Kafka.Addr}
	Topic = conf.Cfg.Kafka.Topic
	Num = conf.Cfg.Kafka.Num

	access.First = time.Now().Unix()
	access.Last = time.Now().Unix()
	access.Count = 0
	var err error

	if err = access.Topic(); err != nil {
		s10 := time.NewTicker(10 * time.Second)
		defer func() {
			s10.Stop()
		}()
		for {
			select {
			case <-s10.C:
				if err = access.Topic(); err == nil {
					break
				}
				log.Printf("kafka get topic err, %v", err)
			}

		}
	}
	//}
	//}

}
