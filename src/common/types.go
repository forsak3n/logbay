package common

const (
	INGEST_TYPE_TLS   IngestType = "tls"
	INGEST_TYPE_REDIS IngestType = "redis"
	INGEST_TYPE_HTTPS IngestType = "https"

	DIGEST_TYPE_REDIS     DigestType = "redis"
	DIGEST_TYPE_WEBSOCKET DigestType = "ws"
	DIGEST_TYPE_FILE      DigestType = "file"
)

type DigestType string
type IngestType string

type AppConfig struct {
	LogConfig    LogConfig              `toml:"Logger"`
	IngestPoints map[string]PointConfig `toml:"IngestPoints"`
	DigestPoints map[string]PointConfig `toml:"DigestPoints"`
}

type LogConfig struct {
	Level  string   `toml:"Level,omitempty"`
	File   string   `toml:"File,omitempty"`
	Fields []string `toml:"Fields,omitempty"`
}

type PointConfig struct {
	Name        string   `toml:"Name"`
	Type        string   `toml:"Type,omitempty"`
	Disabled    bool     `toml:"Disabled,omitempty"`
	Host        string   `toml:"Host,omitempty"`
	Port        int      `toml:"Port,omitempty"`
	Endpoint    string   `toml:"Endpoint,omitempty"`
	Pattern     string   `toml:"Pattern,omitempty"`
	Certificate string   `toml:"Certificate,omitempty"`
	Key         string   `toml:"Key,omitempty"`
	CA          string   `toml:"CA,omitempty"`
	Ingests     []string `toml:"Ingests,omitempty"`
	Delimiter   byte     `toml:"Delimiter,omitempty"`
}

type IngestPoint struct {
	Type IngestType
	Name string
	Msg  chan string
}

type DigestPoint struct {
	Type DigestType
	Name string
}

type Consumer interface {
	Consume(msg chan string) error
}

type Messenger interface {
	Messages() chan string
}
