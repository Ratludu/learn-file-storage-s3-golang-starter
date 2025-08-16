package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
)

type streamsData struct {
	Streams []struct {
		Width              int    `json:"width"`
		Height             int    `json:"height"`
		DisplayAspectRatio string `json:"display_aspect_ratio"`
	} `json:"streams"`
}

func CatAspectRatio(ratio string) string {
	switch ratio {
	case "16:9":
		return "/landscape/"
	case "9:16":
		return "/portrait/"
	default:
		return "/other/"
	}
}

func getVideoAspectRatio(filePath string) (string, error) {

	var stdOutBuffer bytes.Buffer

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	cmd.Stdout = &stdOutBuffer

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}

	var streamsAspect streamsData

	err = json.Unmarshal(stdOutBuffer.Bytes(), &streamsAspect)
	if err != nil {
		log.Fatalf("Failed to unmarshal data: %v", err)
	}

	ratio := streamsAspect.Streams[0].DisplayAspectRatio

	if ratio == "16:9" || ratio == "9:16" {
		return ratio, nil
	}

	return "other", nil
}
