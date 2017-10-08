package main

import (
	"./common"
	"./digest"
	"./ingest"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

var log = common.ContextLogger(context.WithValue(context.Background(), "prefix", "main"))

func main() {

	confPath := flag.String("c", "config.toml", "Specifies config file location. Defaults to config.toml")
	flag.Parse()

	config, err := loadConfig(confPath)

	if err != nil {
		log.Errorf("Can not load config file at %s. Err: %s", confPath, err.Error())
		os.Exit(1)
	}

	prepareLogger(config.LogConfig)
	prepareIngests(config.IngestPoints)
	prepareDigests(config.DigestPoints)

	go runTest()

	<-make(chan byte)
}

func loadConfig(p *string) (*common.AppConfig, error) {
	config := &common.AppConfig{}
	_, err := toml.DecodeFile(*p, config)
	return config, err
}

func prepareIngests(ingests map[string]common.PointConfig) {

	for k, v := range ingests {
		if v.Enabled {
			v.Name = k
			_, err := ingest.NewIngestPoint(v)

			if err != nil {
				logrus.Errorf("Failed to create ingest point. Err: %s", err.Error())
				continue
			}
		}
	}
}

func prepareDigests(digests map[string]common.PointConfig) {
	for k, dp := range digests {
		if dp.Enabled {

			dp.Name = k
			ingests := make([]common.IngestPoint, 0)

			// gather ingest points
			for _, ip := range dp.Ingests {
				if point, ok := ingest.GetIngestPoint(ip); !ok {
					log.Warnf("DigestPoint %s has %s IngestPoint configured, but no such IngestPoint exists", dp.Name, ip)
					continue
				} else {
					ingests = append(ingests, point)
				}
			}

			_, err := digest.NewDigestPoint(dp, ingests)

			if err != nil {
				logrus.Errorf("Failed to create digest point. Err: %s", err.Error())
				continue
			}
		}
	}
}

func prepareLogger(config common.LogConfig) {

	// set log level
	if level := config.Level; len(level) > 0 {
		if level, err := logrus.ParseLevel(level); err != nil {
			log.Errorf("%s is not valid config level. Must be ne of: DEBUG, INFO. WARN, ERROR, FATAL")
		} else {
			logrus.SetLevel(level)
		}
	}

	// set additional logging fields
	if fields := config.Fields; len(fields) > 0 {
		common.SetContextFields(fields)
	}

	// redirect log output to file
	if logFile := config.File; len(logFile) > 0 {

		if dir := filepath.Dir(logFile); dir == "." {

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Errorf("Can't create log directory %s. Err: %s", dir, err.Error())
				return
			}

		} else {
			f, err := rotateLog(logFile)

			if err != nil {
				log.Errorf("Can't write to log file. Err: %s", err.Error())
				return
			}

			logrus.SetOutput(f)
		}
	}
}

func rotateLog(logfile string) (*os.File, error) {

	f, err := os.OpenFile(logfile, os.O_WRONLY, os.ModePerm)
	defer f.Close()

	if err == nil {
		old := fmt.Sprintf("%s.%s", logfile, time.Now().Format("20060102-150405"))
		err = os.Rename(f.Name(), old)

		if err != nil {
			return nil, err
		}
	}

	return os.Create(logfile)
}

func runTest() {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", "127.0.0.1:8001", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	for {
		conn.Write([]byte("{\"companyId\": \"teradek\"}\n"))
		time.Sleep(time.Second * 2)
		conn.Write([]byte("{\"companyId\": \"webb\"}\n"))
		time.Sleep(time.Second * 2)
	}
}
