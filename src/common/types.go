package common

const (
	IngestTLS       IngestType = "tls"
	IngestRedis     IngestType = "redis"
	IngestHTTPS     IngestType = "https"
	IngestSimulated IngestType = "simulated"

	DigestRedis     DigestType = "redis"
	DigestWebSocket DigestType = "ws"
	DigestFile      DigestType = "file"
	DigestElastic   DigestType = "elastic"
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
	Buffer      int      `toml:"Buffer,omitempty"`
	ESIndex     string   `toml:"ESIndex,omitempty"`
	ESDocument  string   `toml:"ESDocument,omitempty"`
	ESBatchSize int      `toml:"ESBatchSize,omitempty"`
	MsgLength   int      `toml:"MsgLength,omitempty"`
	MsgPerSec   int      `toml:"MsgPerSec,omitempty"`
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
	Consume(msg string) error
}

type Messenger interface {
	Messages() chan string
}
