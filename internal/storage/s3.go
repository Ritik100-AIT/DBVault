package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements StorageBackend for AWS S3.
type S3Storage struct {
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	Endpoint       string
	ForcePathStyle bool
	Prefix         string
	client         *s3.Client
}

func (s *S3Storage) initClient(ctx context.Context) error {
	if s.client != nil {
		return nil
	}

	if s.Bucket == "" {
		return errors.New("s3 bucket is not configured")
	}

	loadOptions := []func(*config.LoadOptions) error{}
	region := s.Region
	if region == "" {
		region = "us-east-1"
	}
	loadOptions = append(loadOptions, config.WithRegion(region))

	if s.AccessKey != "" && s.SecretKey != "" {
		creds := credentials.NewStaticCredentialsProvider(s.AccessKey, s.SecretKey, "")
		loadOptions = append(loadOptions, config.WithCredentialsProvider(creds))
	}

	if s.Endpoint != "" {
		parsed, err := url.Parse(s.Endpoint)
		if err != nil {
			return err
		}
		if parsed.Scheme == "" {
			parsed.Scheme = "https"
		}
		endpointURL := parsed.String()
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpointURL,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		})
		loadOptions = append(loadOptions, config.WithEndpointResolverWithOptions(resolver))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return err
	}

	s.client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s.ForcePathStyle
	})
	return nil
}

func (s *S3Storage) buildKey(name string) string {
	cleanPrefix := strings.TrimSpace(s.Prefix)
	if cleanPrefix == "" {
		return name
	}
	return strings.TrimLeft(path.Join(cleanPrefix, name), "/")
}

func parseS3Path(spec string) (bucket, key string, err error) {
	if strings.HasPrefix(spec, "s3://") {
		parsed, parseErr := url.Parse(spec)
		if parseErr != nil {
			return "", "", parseErr
		}
		bucket = parsed.Host
		key = strings.TrimLeft(parsed.Path, "/")
		return bucket, key, nil
	}
	return "", strings.TrimLeft(spec, "/"), nil
}

// Save uploads the given reader to S3 under the configured bucket and prefix.
func (s *S3Storage) Save(name string, src io.Reader) (string, error) {
	if err := s.initClient(context.Background()); err != nil {
		return "", err
	}

	key := s.buildKey(name)
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   src,
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("s3://%s/%s", s.Bucket, key), nil
}

// Load downloads the object from S3 and returns a read closer for its contents.
func (s *S3Storage) Load(spec string) (io.ReadCloser, error) {
	if err := s.initClient(context.Background()); err != nil {
		return nil, err
	}

	bucket, key, err := parseS3Path(spec)
	if err != nil {
		return nil, err
	}
	if bucket == "" {
		bucket = s.Bucket
	}

	resp, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
