package ingest

import (
	"../common"
	"context"
	"errors"
	"fmt"
	"sync"
)

type ingest struct {
	ingestType string
	name       string
	out        chan string
}

func (i *ingest) IngestType() string {
	return i.ingestType
}

func (i *ingest) Name() string {
	return i.name
}

func (i *ingest) Output() chan string {
	return i.out
}

var (
	log     = common.ContextLogger(context.WithValue(context.Background(), "prefix", "launcher"))
	storage = struct {
		mux     *sync.Mutex
		ingests map[string]common.IngestPoint
	}{
		mux:     &sync.Mutex{},
		ingests: make(map[string]common.IngestPoint),
	}
)

func NewIngestPoint(i common.PointConfig) (common.IngestPoint, error) {

	log.Infof("Starting %s ingest point. Type: %s", i.Name, i.Type)

	if _, ok := GetIngestPoint(i.Name); ok {
		return nil, errors.New("already exists")
	}

	switch i.Type {
	case common.INGEST_TYPE_TLS:
		point, err := NewTLSIngest(i.Name, &tlsConfig{
			Port: i.Port,
			Cert: i.Certificate,
			Key:  i.Key,
			CA:   i.CA,
			Delimiter: i.Delimiter,
		})

		if err == nil {
			SetIngestPoint(point)
		}

		return point, err
	case common.INGEST_TYPE_REDIS:
		point, err := NewRedisIngest(i.Name, &redisConf{
			Host:    i.Host,
			Port:    i.Port,
			Channel: i.Pattern,
		})

		if err == nil {
			SetIngestPoint(point)
		}

		return point, err
	case common.INGEST_TYPE_HTTPS:
	}

	return nil, errors.New(fmt.Sprintf("invalid ingest point type %s", i.Type))
}

func GetIngestPoint(name string) (common.IngestPoint, bool) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	point, ok := storage.ingests[name]
	return point, ok
}

func SetIngestPoint(i common.IngestPoint) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	_, ok := storage.ingests[i.Name()]

	if ok {
		log.Warnf("IngestPoint %s already exists", i.Name())
	} else {
		storage.ingests[i.Name()] = i
	}
}
