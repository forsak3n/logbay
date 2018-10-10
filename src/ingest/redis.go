package ingest

import (
	"../common"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
)

type redisConf struct {
	Host    string
	Port    int
	Channel string
}

type redisIngest struct {
	common.IngestPoint
	name string
	pub  *redis.PubSub
	out  chan string
}

func NewRedisIngest(name string, conf *redisConf) (common.IngestPoint, error) {

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

	r := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})

	ingest := &redisIngest{
		name: name,
		pub:  r.PSubscribe(conf.Channel),
		out:  make(chan string),
	}

	go ingest.consume()

	return ingest, nil
}

func (i *redisIngest) IngestType() string {
	return common.INGEST_TYPE_REDIS
}

func (i *redisIngest) Name() string {
	return i.name
}

func (i *redisIngest) Output() chan string {
	return i.out
}

func (i *redisIngest) consume() {
	for {
		select {
		case msg := <-i.pub.Channel():
			i.write(msg.Payload)
		default:
			// do nothing
		}
	}
}

func (i *redisIngest) write(msg string) {

	select {
	case i.out <- msg:
	default:
		// do nothing
	}
}
