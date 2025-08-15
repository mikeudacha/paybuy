package user

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mikeudacha/paybuy/config"
	"github.com/mikeudacha/paybuy/models"
	"github.com/mikeudacha/paybuy/services/auth"
	"github.com/mikeudacha/paybuy/utils"
)

type Handler struct {
	store          models.UserStore
	blacklistStore *auth.BlacklistStore
}

func NewHandler(store models.UserStore, blacklistStore *auth.BlacklistStore) *Handler {
	return &Handler{
		store:          store,
		blacklistStore: blacklistStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/refresh", h.handleRefreshToken).Methods("POST")
	router.HandleFunc("/logout", h.handleLogout).Methods("POST")

	router.HandleFunc("/users/{userID}", auth.WithJWTAuth(h.handleGetUser, h.store, h.blacklistStore)).Methods(http.MethodGet)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var user models.LoginUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validator.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, errors)
		return
	}

	u, err := h.store.GetUserByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("not found, invalid email or password"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(user.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email or password"))
		return
	}

	tokenPair, err := auth.CreateTokenPair(u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	cfg := config.LoadConfig()
	refreshExpiration, _ := strconv.Atoi(cfg.JWTRefreshExpirationInSeconds)
	utils.SetHttpOnlyCookie(w, "refresh_token", tokenPair.RefreshToken, refreshExpiration)

	response := models.LoginResponse{
		AccessToken: tokenPair.AccessToken,
		ExpiresIn:   tokenPair.ExpiresIn,
		Message:     "Successfully logged in. Refresh token stored in httpOnly cookie.",
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user models.RegisterUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validator.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, errors)
		return
	}

	_, err := h.store.GetUserByEmail(user.Email)
	if err == nil {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("user with this email already exists"))
		return
	}

	hashedPassword, err := auth.HashedPassword(user.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateUser(models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, nil)
}

func (h *Handler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := utils.GetCookieValue(r, "refresh_token")
	if refreshToken == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("refresh token not found in cookies"))
		return
	}

	tokenPair, err := auth.RefreshAccessToken(refreshToken, h.blacklistStore)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	cfg := config.LoadConfig()
	refreshExpiration, _ := strconv.Atoi(cfg.JWTRefreshExpirationInSeconds)
	utils.SetHttpOnlyCookie(w, "refresh_token", tokenPair.RefreshToken, refreshExpiration)

	response := models.LoginResponse{
		AccessToken: tokenPair.AccessToken,
		ExpiresIn:   tokenPair.ExpiresIn,
		Message:     "Access token refreshed successfully. New refresh token stored in httpOnly cookie.",
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken := utils.GetCookieValue(r, "refresh_token")
	if refreshToken == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("refresh token not found in cookies"))
		return
	}

	token, err := auth.ValidateJWT(refreshToken, "refresh")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid refresh token"))
		return
	}

	claims := token.Claims.(*auth.JWTClaims)
	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID in token"))
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	err = h.blacklistStore.AddToBlacklist(refreshToken, userID, expiresAt)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to logout"))
		return
	}

	utils.DeleteCookie(w, "refresh_token")

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out. Refresh token removed from cookies and added to blacklist.",
	})
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["userID"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing user ID"))
		return
	}

	userID, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	user, err := h.store.GetUserByID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}
