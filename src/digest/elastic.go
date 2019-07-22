package digest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"logbay/common"
	"math/rand"
	"net/http"
	"time"
)

type ElasticDigestCfg struct {
	Host      string
	Index     string
	Document  string
	BatchSize int
}

type elasticDigest struct {
	common.DigestPoint
	endpoint  string
	index     string
	document  string
	batchSize int
	ch        chan string
}

func (e *elasticDigest) Consume(msg string) error {
	e.ch <- msg
	return nil
}

func (e *elasticDigest) collect() {

	payload := make([]string, e.batchSize)
	counter := 0

loop:
	for {
		select {
		case msg, ok := <-e.ch:

			if !ok {
				send(e.endpoint, payload[:counter])
				break loop
			}

			if counter >= e.batchSize {
				send(e.endpoint, payload)
				counter = 0
			}

			payload[counter] = msg
			counter = counter + 1
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	close(e.ch)
}

func send(endpoint string, strings []string) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "elasticDigest"))

	var buf bytes.Buffer

	for _, v := range strings {
		buf.WriteString(fmt.Sprintf("%s\n", `{ "index" : { } }`))
		buf.WriteString(fmt.Sprintf("%s\n", v))
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(endpoint, "application/x-ndjson", &buf)

	if err != nil {
		log.Errorf("request to %s failed", endpoint, err)
		return
	}

	resp.Body.Close()
}

func NewElasticDigest(name string, cfg *ElasticDigestCfg) (common.Consumer, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "elasticDigest"))

	if len(cfg.Index) == 0 {
		return nil, errors.New("index is required")
	}

	if len(cfg.Document) == 0 {
		return nil, errors.New("document is required")
	}

	if len(cfg.Host) == 0 {
		log.Debugln("Host is not configured. Using localhost")
		cfg.Host = "http://localhost:9200"
	}

	if cfg.BatchSize == 0 {
		log.Debugln("BatchSize is not configured. Using 100")
		cfg.BatchSize = 100
	}

	if len(name) == 0 {
		name = fmt.Sprintf("elastic-digest#%d", rand.Int())
	}

	d := &elasticDigest{
		common.DigestPoint{
			Name: name,
			Type: common.DIGEST_TYPE_ELASTIC,
		},
		fmt.Sprintf("%s/%s/%s/_bulk", cfg.Host, cfg.Index, cfg.Document),
		cfg.Index,
		cfg.Document,
		cfg.BatchSize,
		make(chan string),
	}

	go d.collect()

	return d, nil
}
