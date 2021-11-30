package store

import (
	"context"
	"fmt"
	"github.com/asavt7/my-file-service/internal/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	pingBucket = "ping"
)

type AwsStore struct {
	config *aws.Config

	session *session.Session

	uploader   *s3.S3
	downloader s3manageriface.DownloaderAPI
}

func NewAwsStore(config *aws.Config) (*AwsStore, error) {
	a := &AwsStore{config: config}
	err := a.init()
	if err != nil {
		return a, err
	}
	return a, nil
}

func (a *AwsStore) init() error {
	log.Infof("Initializing store")
	newSession, err := session.NewSession(a.config)
	if err != nil {
		return err
	}
	a.session = newSession
	a.uploader = s3.New(newSession)
	a.downloader = s3manager.NewDownloader(newSession)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = a.initPingBucket(ctx)
	if err != nil {
		return fmt.Errorf("ERROR initialize aws store client : failed create ling bucket %+v", err)
	}
	err = a.checkConn(ctx)
	if err != nil {
		return fmt.Errorf("ERROR initialize aws store client %+v", err)
	}

	return nil
}

func (a *AwsStore) LoadFile(ctx context.Context, f model.FileToDownload) (model.LoadedFile, error) {

	buff := &aws.WriteAtBuffer{}
	numBytes, err := a.downloader.DownloadWithContext(ctx,
		buff,
		&s3.GetObjectInput{
			Bucket: aws.String(f.Bucket),
			Key:    aws.String(f.Name),
		})
	if err != nil {
		log.Errorf("Failed to download file %+v", err)
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey, s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
				return model.LoadedFile{}, model.ErrNotFound
			}
		}
		return model.LoadedFile{}, err
	}
	log.Infof("Downloaded file name=%s size=%d bytes", f.Name, numBytes)

	return model.LoadedFile{
		Name: f.Name,
		Body: buff.Bytes(),
	}, nil
}

func (a *AwsStore) SaveFile(ctx context.Context, f model.FileToStore) (model.StoredFile, error) {
	bucket := aws.String(f.Bucket)
	key := aws.String(f.Name)

	if err := a.createBucketIfNotExists(ctx, bucket); err != nil {
		return model.StoredFile{}, err
	}

	_, err := a.uploader.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   f.Body,
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		log.Errorf("Failed to upload data to %s/%s, %s\n", *bucket, *key, err.Error())
		return model.StoredFile{}, err
	}
	log.Infof("Successfully created bucket %s and uploaded data with key %s\n", *bucket, *key)

	return model.StoredFile{
		Name:   f.Name,
		Bucket: f.Bucket,
	}, nil
}

func (a *AwsStore) createBucketIfNotExists(ctx context.Context, bucket *string) error {
	cparams := &s3.CreateBucketInput{
		Bucket: bucket,
	}
	_, err := a.uploader.CreateBucketWithContext(ctx, cparams)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
			default:
				log.Errorf(aerr.Error())
				return err
			}
		} else {
			log.Errorf(err.Error())
			return err
		}
	}
	return nil
}

func (a *AwsStore) initPingBucket(ctx context.Context) error {
	bucket := aws.String(pingBucket)
	err := a.createBucketIfNotExists(ctx, bucket)
	return err
}

func (a *AwsStore) checkConn(ctx context.Context) error {
	log.Debugln("AWS S3 client checking connection")
	bucket := aws.String(pingBucket)
	_, err := a.uploader.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: bucket,
	})
	return err
}

func (a *AwsStore) ReadinessProbe(ctx context.Context) error {
	return a.checkConn(ctx)
}

func (a *AwsStore) LivenessProbe(ctx context.Context) error {
	return a.checkConn(ctx)
}
