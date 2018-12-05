package digest

import (
	"../common"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
	"regexp"
)

type RedisDigestCfg struct {
	Host    string
	Port    int
	Channel string
}

type redisDigest struct {
	common.DigestPoint
	redis   *redis.Client
	signals map[string]chan int
	channel string
}

func (r *redisDigest) Consume(msg string) error {

	pattern := regexp.MustCompile("{{(.*?)}}")
	hasTemplates := pattern.MatchString(r.channel)

	channel := r.channel

	if hasTemplates {
		channel = r.replaceTemplates(pattern, r.channel, msg)
	}

	if len(channel) > 0 {
		r.redis.Publish(channel, msg)
	}

	return nil
}

func (r *redisDigest) replaceTemplates(regex *regexp.Regexp, channel string, msg string) string {
	return regex.ReplaceAllStringFunc(r.channel, func(match string) string {
		key := regex.FindStringSubmatch(match)[1]
		pattern := regexp.MustCompile(fmt.Sprintf(`\\?"%s\\?":\\?"(.*?)\\?"`, key))

		if v := pattern.FindStringSubmatch(msg); len(v) > 0 {
			return v[1]
		}

		return match
	})
}

func NewRedisDigest(name string, cfg *RedisDigestCfg) (common.Consumer, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "redisDigest"))

	if len(cfg.Host) == 0 {
		log.Debugln("Host is not configured. Using localhost")
		cfg.Host = "localhost"
	}

	if cfg.Port == 0 {
		log.Debugln("Port is not configured. Using 6379")
		cfg.Port = 6379
	}

	if len(cfg.Channel) == 0 {
		log.Debugln("Channel is not configured. Using logbay:entry")
		cfg.Channel = "logbay:entry"
	}

	if len(name) == 0 {
		name = fmt.Sprintf("redis-digest#%d", rand.Int())
	}

	r := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	})

	log.Infof("Created new redis digest point. Host: %s, Port: %d, Channel: %s", cfg.Host, cfg.Port, cfg.Channel)

	d := &redisDigest{
		common.DigestPoint{
			Name: name,
			Type: common.DIGEST_TYPE_REDIS,
		},
		r,
		make(map[string]chan int, 0),
		cfg.Channel,
	}

	return d, nil
}
