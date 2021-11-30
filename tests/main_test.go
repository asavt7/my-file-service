package tests

import (
	"fmt"
	"github.com/asavt7/my-file-service/internal/app"
	"github.com/asavt7/my-file-service/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

const (
	minioKey    = "minioKeyminioKey"
	minioSecret = "minioKeyminioKey"
)

var (
	minioContainerName = "minio-" + uuid.New().String()
)

type MainTestSuite struct {
	suite.Suite

	cfg     *config.Config
	baseurl string

	pool *dockertest.Pool
}

func TestMainTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(MainTestSuite))
}

func (m *MainTestSuite) SetupSuite() {
	log.Debug("SetupSuite")

	m.initConfigs()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.WithError(err).Fatal("Could not connect to docker")
	}
	pool.MaxWait = 30 * time.Second
	m.pool = pool

	m.initMinioContainer()

	m.initApp()

	// todo  wait while apps initialize -- use readiness probe
	time.Sleep(400 * time.Millisecond)
}

func (m *MainTestSuite) initApp() {
	go app.NewApp(m.cfg).Run()
}

func (m *MainTestSuite) TearDownSuite() {
	log.Info("TearDownSuite")

	if err := m.pool.RemoveContainerByName(minioContainerName); err != nil {
		log.Warning(err)
	}
}

func (m *MainTestSuite) initConfigs() {

	m.cfg = &config.Config{
		LoggerConfig: config.LoggerConfig{Level: "debug"},
		ServerConfig: config.ServerConfig{Port: 8080},
		S3Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(
				minioKey,
				minioSecret,
				""),
			Endpoint:         aws.String("http://localhost:9000"),
			Region:           aws.String("us-east-1"),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		},
	}

	m.baseurl = fmt.Sprintf("http://localhost:%d", m.cfg.ServerConfig.Port)
}

func (m *MainTestSuite) initMinioContainer() {
	op := dockertest.RunOptions{
		Name:       minioContainerName,
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			"MINIO_ROOT_USER=" + minioKey,
			"MINIO_ROOT_PASSWORD=" + minioSecret,
		},
		PortBindings: map[dc.Port][]dc.PortBinding{
			"9000/tcp": {{HostPort: "9000"}},
		},
	}
	resource, err := m.pool.RunWithOptions(&op)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	if err := resource.Expire(30); err != nil {
		log.Fatal(err)
	}

	endpoint := fmt.Sprintf("localhost:%s", resource.GetPort("9000/tcp"))
	if err != nil {
		log.Fatalf("Could start container %s", err)
	}
	m.cfg.S3Config.Endpoint = aws.String(endpoint)

	// readiness probe
	if err := m.pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", endpoint)
		log.Infof("trying to check minio %s", url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	log.Infof("Minio test container started")
}
