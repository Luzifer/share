package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func doListen() error {
	http.HandleFunc("/post", simpleFilePost)
	return http.ListenAndServe(cfg.Listen, nil)
}

func simpleFilePost(res http.ResponseWriter, r *http.Request) {
	f, fh, err := r.FormFile("file")
	if err != nil {
		log.WithError(err).Error("Unable to retrieve file from request")
		http.Error(res, "Could not retrieve your file", http.StatusBadRequest)
		return
	}

	url, err := executeUpload(fh.Filename, f, true, "", false)
	if err != nil {
		log.WithError(err).Error("Uploading file from HTTP request failed")
		http.Error(res, "Failed to upload file. For details see the log.", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(url))
}
