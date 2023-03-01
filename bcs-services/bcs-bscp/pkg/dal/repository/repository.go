package repository

import "net/http"

type FileApiType interface {
	DownloadFile(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
}
