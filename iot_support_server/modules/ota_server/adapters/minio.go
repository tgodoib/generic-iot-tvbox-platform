package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/url"
	"ota_server/config"
	"ota_server/model"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	config config.MinioConfig
	client *minio.Client
}

func NewMinioAdapter(config config.MinioConfig) *Minio {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:           credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure:          true,
		TrailingHeaders: true,
	})

	if err != nil {
		log.Fatalln("Error creating minio client:", err)
	}

	exists, err := minioClient.BucketExists(context.Background(), config.BucketName)
	if err != nil {
		log.Fatalln("Error checking bucket:", err)
	}
	if !exists {
		log.Fatalln("Bucket " + config.BucketName + " does not exist")
	}

	fmt.Println("Connected to Minio")

	return &Minio{
		config: config,
		client: minioClient,
	}
}

func (m *Minio) ListFirmware() (model.FirmwareList, error) {
	fws := m.client.ListObjects(context.Background(), m.config.BucketName, minio.ListObjectsOptions{
		Recursive:    true,
		WithMetadata: true,
	})

	list := model.FirmwareList{}

	for fw := range fws {
		if fw.Err != nil {
			return nil, fw.Err
		}

		parts := strings.Split(fw.Key, "/")
		list[parts[0]] = append(list[parts[0]], model.FirmwareListEntry{
			Version:      parts[1],
			LastModified: fw.LastModified,
		})
	}

	return list, nil
}

func (m *Minio) StoreFirmware(firmware model.Firmware) (string, error) {
	hash := sha256.New()
	tee := io.TeeReader(firmware.Bin, hash)

	info, err := m.client.PutObject(
		context.Background(),
		m.config.BucketName,
		fmt.Sprintf("%s/%s", firmware.DeviceClass, firmware.Version),
		tee,
		firmware.Size,
		minio.PutObjectOptions{},
	)

	if err != nil {
		return "", err
	}

	_, err = m.client.CopyObject(
		context.Background(),
		minio.CopyDestOptions{
			Bucket: m.config.BucketName,
			Object: info.Key,
			UserMetadata: map[string]string{
				"Checksum-SHA256": hex.EncodeToString(hash.Sum(nil)),
			},
			ReplaceMetadata: true,
		},
		minio.CopySrcOptions{
			Bucket: m.config.BucketName,
			Object: info.Key,
		},
	)

	if err != nil {
		return "", err
	}

	return info.Key, nil
}

func (m *Minio) DeleteFirmware(id string) error {
	err := m.client.RemoveObject(context.Background(), m.config.BucketName, id, minio.RemoveObjectOptions{})

	if err != nil {
		if minio.ToErrorResponse(err).Code == minio.NoSuchKey {
			return model.ErrFirmwareNotFound
		}
	}

	return err
}

func (m *Minio) GetFirmwareData(id string) (model.FirmwareMeta, error) {
	info, err := m.client.StatObject(context.Background(), m.config.BucketName, id, minio.StatObjectOptions{
		Checksum: true,
	})

	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == minio.NoSuchKey {
			return model.FirmwareMeta{}, model.ErrFirmwareNotFound
		}
		return model.FirmwareMeta{}, err
	}

	checksum := info.UserMetadata["Checksum-Sha256"]
	parts := strings.Split(id, "/")

	return model.FirmwareMeta{
		DeviceClass:  parts[0],
		Version:      parts[1],
		Checksum:     checksum,
		LastModified: info.LastModified,
	}, nil
}

func (m *Minio) GenerateLink(id string) (*url.URL, error) {
	link, err := m.client.PresignedGetObject(context.Background(), m.config.BucketName, id, 24*time.Hour, nil)

	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == minio.NoSuchKey {
			return nil, model.ErrFirmwareNotFound
		}
		return nil, err
	}

	return link, nil
}
