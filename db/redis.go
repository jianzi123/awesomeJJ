package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type RedisClient struct{
	*redis.ClusterClient
}

func NewRedisClient()(*RedisClient, error) {
	var client *redis.ClusterClient
	client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"10.254.20.119:6379", "10.254.22.185:6379",
			"10.254.23.37:6379", "10.254.17.85:6379",
			"10.254.16.104:6379", "10.254.16.218:6379"},
	})
	//client := redis.NewClient(&redis.Options{
	//	Addr:     "10.254.20.119:6379",
	//	Password: "", // no password set
	//	DB:       0,        // use default DB
	//})
	ctx := context.Background()
	client.ConfigSet(ctx, "notify-keyspace-events", "EA")

	redisClient := &RedisClient{client}

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("reis 连接失败：", pong, err)
		return redisClient, err
	}
	fmt.Println("reis 连接成功：", pong)
	go func() {
		timer := time.NewTimer(time.Second * 5)
		//timer.Stop()
		for {
			timer.Reset(time.Second * 5)

			select {
			case <-timer.C:
				pong, err := client.Ping(ctx).Result()
				if err != nil {
					logrus.Errorf("reis 连接失败：", pong, err)
					continue
				}
			}
		}
	}()
	return redisClient, nil
}

func (client *RedisClient)SetKeyTimeout(key, value string, expiration time.Duration)  error{
	// set key value ，并指定 过期时间
	ctx := context.Background()
	isExist, err := client.SetNX(ctx, key, value, expiration).Result()
	if err != nil{
		logrus.Errorf("SetKeyTimeout failed: %v %v", isExist, err)
		return err
	}
	logrus.Infof("SetKeyTimeout suc: %v", isExist)
	return nil
}

func (client *RedisClient)SubscribeCustom(rKey chan<- string) {
	ctx := context.Background()
	pubsub := client.Subscribe(ctx, "__keyevent@0__:expired")

	go func() {
		for {
			logrus.Infof("waiting...")
			msg, err := pubsub.Receive(ctx)
			if err != nil{
				logrus.Errorf("receive loop: %v", err)
				continue
			}
			switch res := msg.(type) {
			case *redis.Message:
				logrus.Infof("Message %s", res.Channel)
				logrus.Infof("Message %s", res.Pattern)
				logrus.Infof("Message %s", res.Payload)
				logrus.Infof("Message %s", res.String())
				// 只会报出key
				rKey<- res.Payload
			case *redis.Subscription:
				logrus.Infof("Subscription %s", res.Channel)
				logrus.Infof("Subscription %s", res.Kind)
				logrus.Infof("Subscription %d", res.Count)
				logrus.Infof("Subscription %s", res.String())
			case error:
				logrus.Errorf("error %v", err)
			case *redis.Pong:
				logrus.Errorf("Pong %s", res.Payload)
				logrus.Errorf("Pong %s", res.String())
			default:
				rType := reflect.TypeOf(msg)
				logrus.Errorf("default pubsub.Receive unknow msg type: %v", rType)
				logrus.Errorf("default unknow message %v", res)
			}
		}
	}()

	//_, err := pubsub.Receive(ctx)
	//if err != nil{
	//	logrus.Errorf("receive loop: %v", err)
	//	return
	//}
	//ch := pubsub.Channel()
	//go func() {
	//	for {
	//		select {
	//		case msg := <-ch:
	//			// Trigger bug: Simulating a very slow consumer
	//			fmt.Sprintf("pubsub.Channel %v", msg)
	//			time.Sleep(time.Second * 50)
	//		}
	//	}
	//}()
	//for msg := range ch {
	//	//fmt.Println( msg.Channel, msg.Payload, "\r\n")
	//	fmt.Sprintf("pubsub.Channel %v", msg)
	//}
}