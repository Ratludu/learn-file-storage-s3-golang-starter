package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	const maxUploadLimit = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadLimit)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	file, _, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not unpack video", err)
		return
	}

	defer file.Close()

	mediaType, _, err := mime.ParseMediaType("video/mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not parse mediaType", err)
		return
	}

	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could create mp4 file", err)
		return
	}

	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err = io.Copy(tempFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}

	tempFile.Seek(0, io.SeekStart)

	ratio, _ := getVideoAspectRatio(tempFile.Name())
	cleanRatio := CatAspectRatio(ratio)

	s3Key, err := makeUniqueName()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making key", err)
		return
	}

	s3Key = cleanRatio + s3Key
	s3Key += ".mp4"

	processTempFileName, err := processVideoForFastStart(tempFile.Name())
	if err != nil {
		log.Fatalf("Could not process video: %v", err)
	}

	processTempFile, err := os.Open(processTempFileName)
	if err != nil {
		log.Fatalf("Could open new video: %v", err)
	}

	defer processTempFile.Close()

	cfg.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &s3Key,
		Body:        processTempFile,
		ContentType: &mediaType,
	})

	s3URL := fmt.Sprintf("https://%s/%s", cfg.s3CfDistribution, s3Key)
	video.VideoURL = &s3URL
	if err != nil {
		log.Fatalf("Could not presign video: %v", err)
	}
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)

}
