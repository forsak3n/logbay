package common

const (
	INGEST_TYPE_TLS   = "tls"
	INGEST_TYPE_REDIS = "redis"
	INGEST_TYPE_HTTPS = "https"

	DIGEST_TYPE_REDIS     = INGEST_TYPE_REDIS
	DIGEST_TYPE_WEBSOCKET = "ws"
	DIGEST_TYPE_FILE      = "file"
)

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
	Enabled     bool     `toml:"Enabled,omitempty"`
	Host        string   `toml:"Host,omitempty"`
	Port        int      `toml:"Port,omitempty"`
	Endpoint    string   `toml:"Endpoint,omitempty"`
	Pattern     string   `toml:"Pattern,omitempty"`
	Certificate string   `toml:"Certificate,omitempty"`
	Key         string   `toml:"Key,omitempty"`
	CA          string   `toml:"CA,omitempty"`
	Ingests     []string `toml:"Ingests,omitempty"`
}

type IngestPoint interface {
	IngestType() string
	Name() string
	Output() chan string
}

type DigestPoint interface {
	DigestType() string
	Name() string
	Ingests() []IngestPoint
	Ingest(name string) (IngestPoint, error)
	AddIngest(i IngestPoint) error
	RemoveIngest(name string)
}
