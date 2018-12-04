package ingest

import (
	"../common"
	"context"
	"errors"
	"fmt"
	"sync"
)

var storage = &sync.Map{}

func NewIngestPoint(i common.PointConfig) (point common.Messenger, err error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "launcher"))

	log.Infof("Starting %s ingest point. Type: %s", i.Name, i.Type)

	if _, ok := GetIngestPoint(i.Name); ok {
		return nil, errors.New("already exists")
	}

	switch common.IngestType(i.Type) {
	case common.INGEST_TYPE_TLS:
		point, err = NewTLSIngest(i.Name, &tlsConfig{
			Port:      i.Port,
			Cert:      i.Certificate,
			Key:       i.Key,
			CA:        i.CA,
			Delimiter: i.Delimiter,
		})
	case common.INGEST_TYPE_REDIS:
		point, err = NewRedisIngest(i.Name, &redisConf{
			Host:    i.Host,
			Port:    i.Port,
			Channel: i.Pattern,
		})
	case common.INGEST_TYPE_HTTPS:
	}

	if err != nil {
		return nil, errors.New(fmt.Sprintf("invalid ingest point type %s", i.Type))
	}

	SetIngestPoint(i.Name, point)
	return point, err
}

func GetIngestPoint(name string) (common.Messenger, bool) {
	value, ok := storage.Load(name)

	if !ok {
		return nil, ok
	}

	return value.(common.Messenger), ok
}

func SetIngestPoint(name string, i common.Messenger) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "launcher"))

	_, ok := storage.Load(name)

	if ok {
		log.Warnf("IngestPoint %s already exists. Skip init", name)
		return
	}

	storage.Store(name, i)
}
