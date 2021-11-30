package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"os"
)

type ServerConfig struct {
	Host string `envconfig:"HTTP_HOST"`
	Port int    `envconfig:"HTTP_PORT" default:"8080"`
}

type LoggerConfig struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
}

type Config struct {
	LoggerConfig LoggerConfig
	ServerConfig ServerConfig
	S3Config     aws.Config
}

type storeConf struct {
	Key         string `envconfig:"STORE_ACCESS_KEY" required:"true"`
	Secret      string `envconfig:"STORE_ACCESS_SECRET"  required:"true"`
	Endpoint    string `envconfig:"STORE_ENDPOINT"  required:"true"`
	SslDisabled bool   `envconfig:"STORE_SSL_DISABLED" default:"true" required:"true"`
}

func Init() (*Config, error) {

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	cfg.ServerConfig.Host, err = getHost(cfg.ServerConfig.Host)
	if err != nil {
		return nil, err
	}

	err = configureLogger(cfg.LoggerConfig)
	if err != nil {
		return nil, err
	}

	cfg.S3Config, err = initAwsStoreConfig()
	if err != nil {
		return nil, err
	}

	logConfigs(cfg)

	return &cfg, nil
}

func getHost(hostname string) (string, error) {
	if hostname == "" {
		return os.Hostname()
	}
	return hostname, nil
}

func configureLogger(cfg LoggerConfig) error {
	level, err := log.ParseLevel(cfg.Level)
	if err != nil {
		log.Errorf("Error occurred while setting log level")
		return err
	}
	log.SetLevel(level)
	return nil
}

func logConfigs(cfg Config) {
	log.Debugf("Configuration loaded : %+v", cfg)
}

func initAwsStoreConfig() (aws.Config, error) {
	var sc storeConf
	err := envconfig.Process("", &sc)
	if err != nil {
		return aws.Config{}, err
	}

	return aws.Config{
		Credentials: credentials.NewStaticCredentials(
			sc.Key,
			sc.Secret,
			""),
		Endpoint:         aws.String(sc.Endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(sc.SslDisabled),
		S3ForcePathStyle: aws.Bool(true),
	}, nil
}
