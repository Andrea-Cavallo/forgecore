// Package rest exposes the storage-service HTTP API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/storage-service/internal/application"
)

const (
	headerTenantID = "X-Tenant-ID"
	headerUserID   = "X-User-ID"
	maxUploadBytes = 100 * 1024 * 1024 // 100 MB
)

// Handler holds all HTTP handlers for the storage service.
type Handler struct {
	upload   *application.UploadUseCase
	presign  *application.GeneratePresignedUseCase
}

// NewHandler constructs the storage REST handler.
func NewHandler(upload *application.UploadUseCase, presign *application.GeneratePresignedUseCase) *Handler {
	return &Handler{upload: upload, presign: presign}
}

// RegisterRoutes mounts storage routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/storage/upload", h.uploadHandler)
	mux.HandleFunc("GET /v1/storage/{id}/presign", h.presignHandler)
}

func (h *Handler) uploadHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractIDs(w, r)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "form multipart non valido")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "campo file mancante")
		return
	}
	defer file.Close()

	out, err := h.upload.Execute(r.Context(), application.UploadInput{
		TenantID:    tenantID,
		UserID:      userID,
		Bucket:      r.FormValue("bucket"),
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out.File)
}

func (h *Handler) presignHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := extractIDs(w, r)
	if !ok {
		return
	}
	fileID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "file id non valido")
		return
	}
	presigned, err := h.presign.Execute(r.Context(), application.GeneratePresignedInput{
		FileID:   fileID,
		TenantID: tenantID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, presigned)
}

// ---- helpers ----

func extractIDs(w http.ResponseWriter, r *http.Request) (tenantID, userID uuid.UUID, ok bool) {
	tenantID, err := uuid.Parse(r.Header.Get(headerTenantID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "tenant_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err = uuid.Parse(r.Header.Get(headerUserID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	return tenantID, userID, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
