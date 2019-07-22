package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"logbay/common"
	"logbay/digest"
	"logbay/ingest"
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
	consumers := prepareDigests(config.DigestPoints)

	dispatch(consumers)

	<-make(chan byte)
}

func loadConfig(p *string) (*common.AppConfig, error) {
	config := &common.AppConfig{}
	_, err := toml.DecodeFile(*p, config)
	return config, err
}

func prepareIngests(ingests map[string]common.PointConfig) {

	for k, v := range ingests {

		if v.Disabled {
			continue
		}

		v.Name = k
		_, err := ingest.NewIngestPoint(v)

		if err != nil {
			logrus.Errorf("Failed to create ingest point. Err: %s", err.Error())
			continue
		}
	}
}

func prepareDigests(conf map[string]common.PointConfig) map[string][]common.Consumer {

	consumers := make(map[string][]common.Consumer)

	for name, pointConfig := range conf {

		if pointConfig.Disabled {
			continue
		}

		pointConfig.Name = name

		consumer, err := digest.New(pointConfig)

		if err != nil {
			logrus.Errorf("Failed to create digest point. Err: %s", err.Error())
			continue
		}

		// check ingest points
		for _, ingestName := range pointConfig.Ingests {
			if _, ok := ingest.GetIngestPoint(ingestName); !ok {
				log.Warnf("DigestPoint %s has %s IngestPoint configured, but no such IngestPoint exists", pointConfig.Name, ingestName)
				continue
			} else {
				log.Debugf("%s consuming from %s", name, ingestName)
				consumers[ingestName] = append(consumers[ingestName], consumer)
			}
		}
	}

	return consumers
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

func dispatch(mapping map[string][]common.Consumer) {

	for ingestName, consumers := range mapping {

		messenger, ok := ingest.GetIngestPoint(ingestName)

		if !ok {
			continue
		}

		go func(m common.Messenger, consumers []common.Consumer) {
			for {
				select {
				case msg := <-messenger.Messages():
					for _, consumer := range consumers {
						consumer.Consume(msg)
					}
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(messenger, consumers)

	}
}
