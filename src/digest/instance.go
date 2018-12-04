package digest

import (
	"../common"
	"context"
	"errors"
	"fmt"
)

//type digest struct {
//	digestType string
//	name       string
//	inputs     []common.Messenger
//	mux        *sync.Mutex
//}

//func (p *digest) DigestType() string {
//	return p.digestType
//}
//
//func (p *digest) Name() string {
//	return p.name
//}
//
//func (p *digest) Ingests() []common.Messenger {
//	return p.inputs
//}
//
//func (p *digest) Ingest(name string) (common.Messenger, error) {
//	for _, v := range p.inputs {
//		if v.Name == name {
//			return v, nil
//		}
//	}
//
//	return nil, errors.New("no such ingest")
//}
//
//func (p *digest) AddIngest(i common.IngestPoint) error {
//	p.mux.Lock()
//	defer p.mux.Unlock()
//
//	for _, v := range p.inputs {
//		if v.Name == i.Name {
//			log.Debugf("Digest %s is already ingesting from %s", p.Name(), i.Name)
//			return errors.New("ingest has already been added")
//		}
//	}
//
//	p.inputs = append(p.inputs, i)
//	return nil
//}
//
//func (p *digest) RemoveIngest(name string) {
//	p.mux.Lock()
//	defer p.mux.Unlock()
//
//	index := -1
//
//	for i, v := range p.inputs {
//		if v.Name == name {
//			log.Debugf("Removing Ingest %s from %s", name, p.Name())
//			index = i
//			break
//		}
//	}
//
//	if index != -1 {
//		p.inputs = append(p.inputs[:index], p.inputs[index+1:]...)
//	}
//}

func New(config common.PointConfig) (common.Consumer, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "digest"))

	log.Infof("Starting %s digest point. Type: %s", config.Name, config.Type)

	switch common.DigestType(config.Type) {
	case common.DIGEST_TYPE_FILE:
	case common.DIGEST_TYPE_REDIS:
		return NewRedisDigest(config.Name, &RedisDigestCfg{
			Host:    config.Host,
			Port:    config.Port,
			Channel: config.Pattern,
		})
	case common.DIGEST_TYPE_WEBSOCKET:
		return NewWSDigest(config.Name, &WSDigestCfg{
			URL:     config.Endpoint,
			Port:    config.Port,
		})
	}

	return nil, errors.New(fmt.Sprintf("Invalid digest point type %s", config.Type))
}
