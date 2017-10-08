package ingest

import (
	"../common"
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type tlsConfig struct {
	Port int
	Cert string
	Key  string
}

func StartTLSServer(name string, config *tlsConfig) (common.IngestPoint, error) {

	log := common.ContextLogger(context.WithValue(context.Background(), "prefix", "tlsIngest"))
	point := &ingest{
		name:       name,
		ingestType: "tls",
		out:        make(chan string),
	}

	if config.Port == 0 {
		log.Warnf("TLS port should be > 0")
		return point, errors.New("invalid port 0")
	}

	if len(config.Cert) == 0 || len(config.Key) == 0 {
		log.Warnf("Invalid certificate or key path. Cert: %s. Key: %s", config.Cert, config.Key)
		return point, errors.New("invalid certificate or key path")
	}

	cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)

	if err != nil {
		log.Errorf("Failed to load keypair. Err: %s", err.Error())
		return point, err
	}

	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
	tlsConfig.Rand = rand.Reader

	server, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port), &tlsConfig)

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

			log.Infof("Accepted connection from %s", conn.RemoteAddr())

			go read(conn, ch)
		}
	}(server, point.Output())

	return point, nil
}

func read(conn net.Conn, ch chan string) {

	defer conn.Close()

	for {
		line, err := bufio.NewReader(conn).ReadString('\n')

		if err != nil && err != io.EOF {
			log.Infof("Socket read error: %s. Terminating connection...", conn.RemoteAddr(), err.Error())
			break
		}

		ch <- strings.TrimSpace(strings.TrimSuffix(line, "\n"))
	}

	log.Infof("Connection from %s closed", conn.RemoteAddr())
}
