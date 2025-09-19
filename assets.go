package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s.%s", id, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}

	return parts[1]
}

func getVideoAspectRatio(filepath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	outBuffer := &bytes.Buffer{}
	cmd.Stdout = outBuffer
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type videoAspect struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	aspect := videoAspect{}
	err = json.Unmarshal(outBuffer.Bytes(), &aspect)
	if err != nil {
		return "", err
	}

	if len(aspect.Streams) < 1 {
		return "", fmt.Errorf("Could not extract aspect from video")
	}

	w := float64(aspect.Streams[0].Width)
	h := float64(aspect.Streams[0].Height)
	log.Printf("w=%v h=%v diff16x9=%v diff9x16=%v",
		w, h, math.Abs(w*9-h*16), math.Abs(w*16-h*9))
	if math.Abs(w*9-h*16) < 100 {
		return "16:9", nil
	}
	if math.Abs(w*16-h*9) < 100 {
		return "9:16", nil
	}
	return "other", nil

}
