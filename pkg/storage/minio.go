package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"hopSpotAPI/internal/config"
)

type MinioClient struct {
	client         *minio.Client
	bucketName     string
	publicEndpoint string
	publicSSL      bool
}

func NewMinioClient(cfg config.Config) (*MinioClient, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Public Endpoint fallback to internal if not set
	publicEndpoint := cfg.MinioPublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = cfg.MinioEndpoint
	}

	return &MinioClient{
		client:         client,
		bucketName:     cfg.MinioBucketName,
		publicEndpoint: publicEndpoint,
		publicSSL:      cfg.MinioPublicSSL,
	}, nil
}

func (m *MinioClient) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}
	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}
	return nil
}

func (m *MinioClient) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

func (m *MinioClient) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return obj, nil
}

func (m *MinioClient) Delete(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for accessing the object directly.
func (m *MinioClient) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}

	// Replace internal endpoint with public endpoint
	urlStr := url.String()
	internalHost := m.client.EndpointURL().Host

	// Build public URL prefix
	publicScheme := "http"
	if m.publicSSL {
		publicScheme = "https"
	}

	// Replace internal with public endpoint (http and https)
	urlStr = strings.Replace(urlStr,
		fmt.Sprintf("http://%s", internalHost),
		fmt.Sprintf("%s://%s", publicScheme, m.publicEndpoint),
		1)
	urlStr = strings.Replace(urlStr,
		fmt.Sprintf("https://%s", internalHost),
		fmt.Sprintf("%s://%s", publicScheme, m.publicEndpoint),
		1)

	return urlStr, nil
}

// GetPublicURL returns the public URL of the object. If the bucket is public.
func (m *MinioClient) GetPublicURL(objectName string) string {
	scheme := "http"
	if m.publicSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, m.publicEndpoint, m.bucketName, objectName)
}
