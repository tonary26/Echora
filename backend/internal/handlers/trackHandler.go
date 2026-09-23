package handlers

import (
	"database/sql"
	"echora/internal/database"
	"echora/internal/models"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TrackHandler struct {
	tracks *database.TrackStore
	files  *database.FileStore
}

func NewTrackHandler(tracks *database.TrackStore, files *database.FileStore) *TrackHandler {
	return &TrackHandler{tracks: tracks, files: files}
}

func (h *TrackHandler) GetUserTracks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Требуется вход.")
		return
	}

	tracks, err := h.tracks.GetUserTracks(ctx, userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Ошибка базы")
		return
	}
	for i := range tracks {
		if tracks[i].Cover_key == nil {
			continue
		}

		coverURL, err := h.files.PresignGet(ctx, *tracks[i].Cover_key, time.Hour)
		if err == nil {
			tracks[i].Cover_url = coverURL
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"tracks": tracks})
}

func (h *TrackHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Требуется вход.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	file, hdr, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Не удалось прочитать файл.")
		return
	}
	defer file.Close()

	audioKey := fmt.Sprintf("tracks/%s/%s%s", userID, uuid.NewString(), filepath.Ext(hdr.Filename))
	err = h.files.PutFile(ctx, audioKey, file, hdr.Size, hdr.Header.Get("Content-Type"))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Ошибка загрузки файла.")
		return
	}

	var coverKey *string
	coverFile, coverHdr, err := r.FormFile("cover")
	if err == nil {
		defer coverFile.Close()

		contentType := coverHdr.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			_ = h.files.DeleteFile(ctx, audioKey)
			respondWithError(w, http.StatusBadRequest, "Ошибка базы.")
			return
		}

		key := fmt.Sprintf("covers/%s/%s%s", userID, uuid.NewString(), filepath.Ext(coverHdr.Filename))
		if err := h.files.PutFile(ctx, key, coverFile, coverHdr.Size, contentType); err != nil {
			respondWithError(w, http.StatusBadRequest, "Ошибка загрузки облокжи.")
			_ = h.files.DeleteFile(ctx, audioKey)
			return
		}

		coverKey = &key
	}

	track, err := h.tracks.Create(ctx, models.TrackCreateInput{
		User_id:    userID,
		Title:      hdr.Filename,
		Cover_key:  coverKey,
		Object_key: audioKey,
	})
	if err != nil {
		_ = h.files.DeleteFile(ctx, audioKey)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if track.Cover_key != nil {
		coverURL, err := h.files.PresignGet(ctx, *track.Cover_key, time.Hour)
		if err == nil {
			track.Cover_url = coverURL
		}
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{"track": track})
}

func (h *TrackHandler) StreamURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Требуется вход.")
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некоректный ID трека.")
		return
	}

	track, err := h.tracks.GetTrackByID(ctx, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "Трек не найден.")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Ошибка базы")
		return
	}

	url, err := h.files.PresignGet(ctx, track.Object_key, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Не удалось создать ссылку.")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"url": url})
}
