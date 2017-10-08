package digest

import (
	"../common"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
	"regexp"
	"strings"
	"sync"
)

type RedisDigestCfg struct {
	Name    string
	Host    string
	Port    int
	Channel string
	Ingests []common.IngestPoint
}

type redisDigest struct {
	digest
	redis   *redis.Client
	signals map[string]chan int
	channel string
	regex   *regexp.Regexp
}

func (p *redisDigest) AddIngest(i common.IngestPoint) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, v := range p.inputs {
		if v.Name() == i.Name() {
			log.Debugf("Digest %s is already ingesting from %s", p.Name(), i.Name())
			return errors.New("ingest already has been added")
		}
	}

	p.inputs = append(p.inputs, i)
	p.signals[i.Name()] = make(chan int)
	go p.consume(i)
	return nil
}

func (p *redisDigest) RemoveIngest(name string) {
	p.mux.Lock()
	defer p.mux.Unlock()

	index := -1

	for i, v := range p.inputs {
		if v.Name() == name {
			log.Debugf("Removing Ingest %s from %s", name, p.Name())
			index = i
			break
		}
	}

	if index != -1 {
		p.signals[name] <- -1
		close(p.signals[name])
		p.inputs = append(p.inputs[:index], p.inputs[index+1:]...)
	}
}

func (p *redisDigest) consume(i common.IngestPoint) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "redisDigest"))
	log.Infof("Start consuming from %s", i.Name())

	for {
		select {
		case msg := <-i.Output():
			p.redis.Publish(p.expand(p.channel, msg), msg)
		case <-p.signals[i.Name()]:
			log.Infof("Got signal. Stop consuming from %s", i.Name())
			break
		}
	}
}

func (p *redisDigest) expand(channel string, msg string) string {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "redisDigest"))

	m := make(map[string]string)
	err := json.Unmarshal([]byte(msg), &m)

	if err != nil {
		log.Errorf("Failed to unmarshal message. Raw: %s. Err: %s", msg, err.Error())
		return channel
	}

	return p.regex.ReplaceAllStringFunc(p.channel, func(match string) string {
		clean := strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}")
		if v, ok := m[clean]; ok {
			return v
		}
		return match
	})
}

func NewRedisDigest(cfg *RedisDigestCfg) (common.DigestPoint, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "redisDigest"))
	log.Debugf("Creating new redis digest point. Host: %s, Port: %d, Channel: %s", cfg.Host, cfg.Port, cfg.Channel)

	if cfg.Port == 0 {
		cfg.Port = 6379
	}

	if len(cfg.Host) == 0 {
		cfg.Host = "0.0.0.0"
	}

	if len(cfg.Channel) == 0 {
		cfg.Channel = "logthing:entry"
	}

	if len(cfg.Name) == 0 {
		cfg.Name = fmt.Sprintf("redis-digest#%d", rand.Int())
	}

	redis := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	})

	regex, _ := regexp.Compile(`{{(.*?)}}`)

	log.Infof("DigestPoint %s created", cfg.Name)

	d := &redisDigest{
		digest{
			mux:        &sync.Mutex{},
			name:       cfg.Name,
			digestType: "redis",
			inputs:     cfg.Ingests,
		},
		redis,
		make(map[string]chan int, 0),
		cfg.Channel,
		regex,
	}

	for _, v := range cfg.Ingests {
		go d.consume(v)
	}

	return d, nil
}
