package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(videoURL string, mediaType string) string {
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", videoURL, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getMimeMediaType(header *multipart.FileHeader) error {
	mediaType := header.Header.Get("Content-Type")
	compatibleTypes := []string{"image/png", "image/jpeg"}
	if slices.Contains(compatibleTypes, mediaType) {
		return nil
	}

	return errors.New("The media type is not compatible")
}

func makeUniqueName() (string, error) {

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(key), nil
}
