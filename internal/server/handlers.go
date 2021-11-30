package server

import (
	"bytes"
	"fmt"
	// import swagger specs
	_ "github.com/asavt7/my-file-service/api"
	"github.com/asavt7/my-file-service/internal/model"
	"github.com/google/uuid"
	"github.com/swaggo/http-swagger"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	healthReadiness = "/api/health/readiness"
	healthLiveness  = "/api/health/liveness"
	uploadPath      = "/api/upload"
	downloadPath    = "/api/download/"
)

func (s *APIServer) initHandlers() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc(healthReadiness, s.readiness)
	r.HandleFunc(healthLiveness, s.liveness)

	r.HandleFunc(uploadPath, s.uploadFile)
	r.HandleFunc(downloadPath+"{fileId}", s.downloadFile)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	return s.timeoutMiddleware(r)
}

// uploadFile
// @Summary uploadFile
// @Description upload image of jpeg,png types
// @Tags files
// @Accept multipart/form-data
// @Produce text/html
// @Param file formData file true "file to upload"
// @Success 200
// @Router /api/upload [post]
func (s *APIServer) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error Retrieving the File", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filetype, err := checkFileType(w, err, file)
	if err != nil {
		return
	}

	fileName := generateName(filetype)

	//todo add progress bar like https://freshman.tech/file-upload-golang/
	saveFile, err := s.store.SaveFile(r.Context(), model.FileToStore{
		Body:   file,
		Name:   fileName,
		Bucket: model.AvatarsBucket,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", s.getLocation(saveFile))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	return
}

func generateName(filetype string) string {
	ext := getFileExtension(filetype)
	return uuid.New().String() + strconv.Itoa(int(time.Now().Unix())) + ext
}

func getFileExtension(filetype string) string {
	if filetype == "image/jpeg" {
		return ".jpeg"
	}
	if filetype == "image/png" {
		return ".png"
	}
	return ""
}

func checkFileType(w http.ResponseWriter, err error, file multipart.File) (string, error) {
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		http.Error(w, "The provided file format is not allowed. Please upload a JPEG or PNG image", http.StatusBadRequest)
		return "", err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}
	return filetype, nil
}

func (s *APIServer) getLocation(file model.StoredFile) string {
	return fmt.Sprintf("http://%s:%d/api/download/%s", s.config.Host, s.config.Port, file.Name)
}

// downloadFile
// @Summary downloadFile
// @Description download image of jpeg,png types
// @Tags files
// @Accept multipart/form-data
// @Produce image/png
// @Produce image/jpeg
// @Param fileId path string true "fileId"
// @Success 200
// @Router /api/download/{fileId} [get]
func (s *APIServer) downloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	vars := mux.Vars(r)
	imageID, ok := vars["fileId"]
	if !ok {
		http.Error(w, "Path param fileId not provided", http.StatusBadRequest)
		return
	}

	loadedFile, err := s.store.LoadFile(r.Context(), model.FileToDownload{
		Name:   imageID,
		Bucket: model.AvatarsBucket,
	})
	if err != nil {
		if err == model.ErrNotFound {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, loadedFile.Name, time.Now(), bytes.NewReader(loadedFile.Body))
}
