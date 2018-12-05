package ingest

import (
	"../common"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type simulatorConf struct {
	MsgLength int
	MsgPerSec int
	Buffer    int
}

type simulatedIngest struct {
	msgLength  int
	throughput int
	msg        chan string
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func NewSimulatedIngest(name string, conf *simulatorConf) (common.Messenger, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "simulatedIngest"))

	if len(name) == 0 {
		name = fmt.Sprintf("simulated-ingest#%d", rand.Int())
	}

	if conf.MsgLength == 0 {
		log.Debugln("MsgLength is not defined. Using 140")
		conf.MsgLength = 140
	}

	if conf.MsgPerSec == 0 {
		log.Debugln("MsgPerSec is not defined. Using 5")
		conf.MsgPerSec = 5
	}

	if conf.Buffer == 0 {
		conf.Buffer = 50
	}

	s := &simulatedIngest{
		msgLength:  conf.MsgLength,
		throughput: conf.MsgPerSec,
		msg:        make(chan string, conf.Buffer),
	}

	go s.start()

	return s, nil
}

func (i *simulatedIngest) Messages() chan string {
	return i.msg
}

func (i *simulatedIngest) start() {

	intervalMs := 60 * 1000 / (i.throughput * 60)
	randomDelay := rand.Intn(5000)

	timer := time.NewTicker(time.Duration(intervalMs+randomDelay) * time.Microsecond)

	for range timer.C {
		i.msg <- String(i.msgLength)
	}
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
