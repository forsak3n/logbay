package digest

import (
	"../common"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
	"regexp"
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
}

func (p *redisDigest) AddIngest(i common.IngestPoint) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, v := range p.inputs {
		if v.Name() == i.Name() {
			log.Debugf("Digest %s is already ingesting from %s. Check configuration.", p.Name(), i.Name())
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
	log.Infof("%s started consuming from %s", p.Name(), i.Name())

	r := regexp.MustCompile("{{(.*?)}}")
	hasTemplates := r.MatchString(p.channel)

loop:
	for {
		select {
		case msg, ok := <-i.Output():

			if !ok {
				break loop
			}

			var channel string

			if hasTemplates {
				channel = p.replaceTemplates(r, p.channel, msg)
			}

			if len(channel) > 0 {
				p.redis.Publish(channel, msg)
			}

		case <-p.signals[i.Name()]:
			log.Infof("Got signal. Stop consuming from %s", i.Name())
			break loop
		}
	}
}

func (p *redisDigest) replaceTemplates(r *regexp.Regexp, channel string, msg string) string {
	return r.ReplaceAllStringFunc(p.channel, func(match string) string {
		key := r.FindStringSubmatch(match)[1]
		pattern := regexp.MustCompile(fmt.Sprintf(`\\?"%s\\?":\\?"(.*?)\\?"`, key))

		if v := pattern.FindStringSubmatch(msg); len(v) > 0 {
			return v[1]
		}

		return match
	})
}

func NewRedisDigest(cfg *RedisDigestCfg) (common.DigestPoint, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "redisDigest"))

	if len(cfg.Host) == 0 {
		log.Debugln("Host is not configured. Using localhost")
		cfg.Host = "localhost"
	}

	if cfg.Port == 0 {
		log.Debugln("Port is not configured. Using localhost")
		cfg.Port = 6379
	}

	if len(cfg.Channel) == 0 {
		log.Debugln("Channel is not configured. Using logbay:entry")
		cfg.Channel = "logbay:entry"
	}

	if len(cfg.Name) == 0 {
		cfg.Name = fmt.Sprintf("redis-digest#%d", rand.Int())
	}

	r := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	})

	log.Infof("Created new redis digest point. Host: %s, Port: %d, Channel: %s", cfg.Host, cfg.Port, cfg.Channel)

	d := &redisDigest{
		digest{
			mux:        &sync.Mutex{},
			name:       cfg.Name,
			digestType: common.DIGEST_TYPE_REDIS,
			inputs:     cfg.Ingests,
		},
		r,
		make(map[string]chan int, 0),
		cfg.Channel,
	}

	for _, v := range cfg.Ingests {
		go d.consume(v)
	}

	return d, nil
}
