package handlers

import (
	"echora/internal/database"
	"echora/internal/models"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	store *database.AuthStore
}

func NewAuthHandler(store *database.AuthStore) *AuthHandler {
	return &AuthHandler{store: store}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.UserRegisterInput
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некоректные данные.")
		return
	}

	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" {
		respondWithError(w, http.StatusBadRequest, "Все поля должны быть заполнены.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка хеширования пароля.")
		return
	}

	createInput := models.UserCreateInput{
		Name:       input.Name,
		Email:      input.Email,
		Avatar_url: input.Avatar_url,
		Password:   string(hash),
	}

	user, err := h.store.Register(ctx, createInput)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Не удалось создать пользователя.")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{"user": user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.UserLoginInput
	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некоректные данные.")
		return
	}

	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.Password) == "" {
		respondWithError(w, http.StatusBadRequest, "Все поля должны быть заполнены.")
		return
	}

	user, err := h.store.GetByEmail(ctx, input.Email)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный email или пароль.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password_hash), []byte(input.Password)); err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный email или пароль.")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Ошибка генерации токена.")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokenString,
		"user":  user,
	})
}

func (h *AuthHandler) CurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Требуется вход.")
		return
	}

	user, err := h.store.GetByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Пользователь не найден.")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}
