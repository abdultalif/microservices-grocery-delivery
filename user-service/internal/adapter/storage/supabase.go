package storage

import (
	"io"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"

	"github.com/labstack/gommon/log"
	storage_go "github.com/supabase-community/storage-go"
)

type SupabaseInterface interface {
	UploadFile(path string, file io.Reader) (string, error)
}

type SupabaseStruct struct {
	cfg *config.Config
}

// UploadFile implements SupabaseInterface.
func (s *SupabaseStruct) UploadFile(path string, file io.Reader) (string, error) {
	client := storage_go.NewClient(
		s.cfg.Storage.URL,
		s.cfg.Storage.Key,
		map[string]string{"Content-Type": "image/png"})

	log.Infof("Uploading to Supabase. Bucket: %s, Path: %s", s.cfg.Storage.Bucket, path)

	uploadResp, err := client.UploadFile(s.cfg.Storage.Bucket, path, file)
	log.Infof("Upload Response: %+v", uploadResp)
	if err != nil {
		log.Errorf("Error upload file to Supabase. Bucket: %s, Path: %s, Error: %v", s.cfg.Storage.Bucket, path, err)
		return "", err
	}

	result := client.GetPublicUrl(s.cfg.Storage.Bucket, path)
	log.Infof("Public URL: %+v", result)

	return result.SignedURL, nil
}

func NewSupabase(cfg *config.Config) SupabaseInterface {
	return &SupabaseStruct{cfg: cfg}
}
