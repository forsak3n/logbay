package digest

import (
	"../common"
	"context"
	"errors"
	"fmt"
	"sync"
)

var log = common.ContextLogger(context.WithValue(context.Background(), "prefix", "digest"))

type digest struct {
	digestType string
	name       string
	inputs     []common.IngestPoint
	mux        *sync.Mutex
}

func (p *digest) DigestType() string {
	return p.digestType
}

func (p *digest) Name() string {
	return p.name
}

func (p *digest) Ingests() []common.IngestPoint {
	return p.inputs
}

func (p *digest) Ingest(name string) (common.IngestPoint, error) {
	for _, v := range p.inputs {
		if v.Name() == name {
			return v, nil
		}
	}

	return nil, errors.New("no such ingest")
}

func (p *digest) AddIngest(i common.IngestPoint) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, v := range p.inputs {
		if v.Name() == i.Name() {
			log.Debugf("Digest %s is already ingesting from %s", p.Name(), i.Name())
			return errors.New("ingest has already been added")
		}
	}

	p.inputs = append(p.inputs, i)
	return nil
}

func (p *digest) RemoveIngest(name string) {
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
		p.inputs = append(p.inputs[:index], p.inputs[index+1:]...)
	}
}

func NewDigestPoint(config common.PointConfig, ingests []common.IngestPoint) (common.DigestPoint, error) {

	log.Infof("Starting %s digest point. Type: %s", config.Name, config.Type)

	switch config.Type {
	case common.DIGEST_TYPE_FILE:
	case common.DIGEST_TYPE_REDIS:
		return NewRedisDigest(&RedisDigestCfg{
			Name:    config.Name,
			Host:    config.Host,
			Port:    config.Port,
			Channel: config.Pattern,
			Ingests: ingests,
		})
	case common.DIGEST_TYPE_WEBSOCKET:
	}

	return nil, errors.New(fmt.Sprintf("Invalid ingest point type %s", config.Type))
}
