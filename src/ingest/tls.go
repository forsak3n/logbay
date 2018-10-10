package ingest

import (
	"../common"
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
)

type tlsConfig struct {
	Port      int
	Cert      string
	Key       string
	CA        string
	Delimiter byte
}

func NewTLSIngest(name string, conf *tlsConfig) (common.IngestPoint, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "tlsIngest"))
	point := &ingest{
		name:       name,
		ingestType: common.INGEST_TYPE_TLS,
		out:        make(chan string),
	}

	if conf.Port == 0 {
		log.Warnf("TLS port should be > 0")
		return point, errors.New("invalid port 0")
	}

	if len(conf.Cert) == 0 || len(conf.Key) == 0 {
		log.Warnf("Invalid certificate or key path. Cert: %s. Key: %s", conf.Cert, conf.Key)
		return point, errors.New("invalid certificate or key path")
	}

	if conf.Delimiter == 0 {
		log.Infof("Delimiter is not configured. Using '\n'")
		conf.Delimiter = '\n'
	}

	cert, err := tls.LoadX509KeyPair(conf.Cert, conf.Key)

	if err != nil {
		log.Errorf("Failed to load keypair. Err: %s", err.Error())
		return point, err
	}

	ca, err := ioutil.ReadFile(conf.CA)

	if err != nil {
		log.Errorf("failed to read root certificate. Err: %s", err.Error())
		return nil, err
	}

	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(ca)

	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: roots}
	tlsConfig.Rand = rand.Reader

	server, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conf.Port), &tlsConfig)

	if err != nil {
		log.Errorf("Failed to start server. Err: %s", err.Error())
		return point, err
	}

	log.Infof("TLS server started. Waiting for connections...")

	go func(server net.Listener, ch chan string) {
		for {
			conn, err := server.Accept()

			if err != nil {
				log.Errorf("Can't accept incoming connection. Err: %s", err.Error())
				continue
			}

			log.Debugf("Accepted connection from %s", conn.RemoteAddr())

			go read(conn, ch, conf.Delimiter)
		}
	}(server, point.Output())

	return point, nil
}

func read(conn net.Conn, ch chan string, delim byte) {

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		b, err := reader.ReadBytes(delim)

		if err != nil && err != io.EOF {
			break
		}

		if len(b) == 0 {
			continue
		}

		select {
		case ch <- string(b[:len(b)-1]):
		default:
			// drop message if there are no consumers
		}

		ch <- string(b[:len(b)-1])
	}

	log.Debugf("Connection from %s has been closed", conn.RemoteAddr())
}
