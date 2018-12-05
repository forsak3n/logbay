package digest

import (
	"../common"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	WRITE_TIMEOUT_SEC  = 10 * time.Second
	MAX_WRITE_ATTEMPTS = 10 * time.Second
)

type WSDigestCfg struct {
	Port    int
	URL     string
	Ingests []common.IngestPoint
}

type wsDigest struct {
	common.DigestPoint
	signals  map[string]chan int
	messages chan *websocket.PreparedMessage
	clients  *sync.Map
}

type wsConn struct {
	conn           *websocket.Conn
	pongReceivedAt time.Time
}

var upgrader = websocket.Upgrader{
	// default options
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewWSDigest(name string, conf *WSDigestCfg) (common.Consumer, error) {

	if conf.Port == 0 {
		return nil, errors.New("port is not defined")
	}

	if len(conf.URL) == 0 {
		conf.URL = "/logbay"
	}

	if conf.URL[0] != '/' {
		conf.URL = fmt.Sprintf("/%s", conf.URL)
	}

	if len(name) == 0 {
		name = fmt.Sprintf("ws-digest#%d", rand.Int())
	}

	d := &wsDigest{
		common.DigestPoint{
			Name: name,
			Type: common.DIGEST_TYPE_WEBSOCKET,
		},
		make(map[string]chan int),
		make(chan *websocket.PreparedMessage),
		&sync.Map{},
	}

	d.listen(conf.URL, conf.Port)

	go d.broadcast()

	return d, nil
}

func (w *wsDigest) listen(url string, port int) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "wsDigest"))

	h := func(rw http.ResponseWriter, r *http.Request) {

		c, err := upgrader.Upgrade(rw, r, nil)

		if err != nil {
			log.Errorln("switching protocols:", err)
			return
		}

		clientKey := c.RemoteAddr().String()
		c.SetCloseHandler(func(code int, text string) error {
			w.clients.Delete(clientKey)
			return nil
		})

		c.SetPongHandler(func(appData string) error {
			value, ok := w.clients.Load(clientKey)

			if !ok {
				return nil
			}

			wsConn, ok := value.(*wsConn)

			if !ok {
				return nil
			}

			wsConn.pongReceivedAt = time.Now();
			w.clients.Store(clientKey, wsConn)
			return nil
		})

		w.clients.Store(clientKey, &wsConn{
			conn: c,
		})
	}

	http.HandleFunc(url, h)
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
}

func (w *wsDigest) Consume(msg string) error {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "wsDigest"))

	m, err := websocket.NewPreparedMessage(websocket.TextMessage, []byte(msg))

	if err != nil {
		log.Debugf("Can't prepare message %v", msg)
		return err
	}

	w.messages <- m

	return nil
}

func (w *wsDigest) broadcast() {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "wsDigest"))

loop:
	for {
		select {
		case msg, ok := <-w.messages:

			if !ok {
				break loop
			}

			w.clients.Range(func(key, value interface{}) bool {

				c := value.(*wsConn)

				if err := c.conn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT_SEC)); err != nil {
					log.Errorf("set write deadline failed on %s", c.conn.RemoteAddr().String())
					return true
				}

				if netErr, ok := c.conn.WritePreparedMessage(msg).(net.Error); ok && netErr.Timeout() {
					log.Errorln("write timeout. addr: ", c.conn.RemoteAddr().String())
				}

				return true

			})

		default:
			// do nothing
		}
	}
}
