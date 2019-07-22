package ingest

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"logbay/common"
	"math/rand"
	"time"
)

type redisConf struct {
	Host    string
	Port    int
	Channel string
	Buffer  int
}

type redisIngest struct {
	common.IngestPoint
	pub *redis.PubSub
}

func NewRedisIngest(name string, conf *redisConf) (common.Messenger, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "redisIngest"))

	var host, port = conf.Host, conf.Port

	if len(name) == 0 {
		name = fmt.Sprintf("redis-ingest#%d", rand.Int())
	}

	if len(conf.Channel) == 0 {
		return nil, errors.New("channel can not be empty")
	}

	if len(conf.Host) == 0 {
		log.Debugln("Host is not configured. Using localhost")
		host = "localhost"
	}

	if conf.Port == 0 {
		log.Debugln("Port is not configured. Using 6379")
		port = 6379
	}

	if conf.Buffer == 0 {
		conf.Buffer = 50
	}

	r := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})

	ingest := &redisIngest{
		common.IngestPoint{
			Type: common.INGEST_TYPE_REDIS,
			Name: name,
			Msg:  make(chan string, conf.Buffer),
		},
		r.PSubscribe(conf.Channel),
	}

	go ingest.read()

	return ingest, nil
}

func (i *redisIngest) Messages() chan string {
	return i.Msg
}

func (i *redisIngest) read() {
	for {
		select {
		case msg := <-i.pub.Channel():
			i.write(msg.Payload)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (i *redisIngest) write(msg string) {

	select {
	case i.Msg <- msg:
	default:
		// drop message if there are no consumers or if channel buffer is full. wait a little to reduce steal time
		time.Sleep(100 * time.Millisecond)
	}
}
