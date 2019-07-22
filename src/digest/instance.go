package digest

import (
	"context"
	"errors"
	"fmt"

	"logbay/common"
)

func New(config common.PointConfig) (common.Consumer, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "digest"))

	log.Infof("Starting %s digest point. Type: %s", config.Name, config.Type)

	switch common.DigestType(config.Type) {
	case common.DigestElastic:
		return NewElasticDigest(config.Name, &ElasticDigestCfg{
			Host:      config.Host,
			Index:     config.ESIndex,
			Document:  config.ESDocument,
			BatchSize: config.ESBatchSize,
		})
	case common.DigestRedis:
		return NewRedisDigest(config.Name, &RedisDigestCfg{
			Host:    config.Host,
			Port:    config.Port,
			Channel: config.Pattern,
		})
	case common.DigestWebSocket:
		return NewWSDigest(config.Name, &WSDigestCfg{
			URL:  config.Endpoint,
			Port: config.Port,
		})
	}

	return nil, errors.New(fmt.Sprintf("Invalid digest point type %s", config.Type))
}
