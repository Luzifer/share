package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func doListen() (err error) {
	http.HandleFunc("/post", simpleFilePost)

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: time.Second,
	}

	if err = server.ListenAndServe(); err != nil {
		return fmt.Errorf("listening for HTTP traffic: %w", err)
	}

	return nil
}

func simpleFilePost(w http.ResponseWriter, r *http.Request) {
	var (
		reqUUID = uuid.Must(uuid.NewV4()).String()
		logger  = logrus.WithField("req-id", reqUUID)
		errStr  = fmt.Sprintf("something went wrong: %s", reqUUID)
	)

	f, fh, err := r.FormFile("file")
	if err != nil {
		logger.WithError(err).Error("retrieving file from request")
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	url, err := executeUpload(fh.Filename, f, true, "", false)
	if err != nil {
		logger.WithError(err).Error("uploading file from HTTP request")
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	http.Error(w, url, http.StatusOK)
}
