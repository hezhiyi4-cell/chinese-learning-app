package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"chinese-learning-app/internal/middleware"
	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"
	"chinese-learning-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

type ToneBattleHandler struct {
	toneBattleService *services.ToneBattleService
	userRepo          *repositories.UserRepository
	jwtSecret         string
	upgrader          websocket.Upgrader
}

func NewToneBattleHandler(
	toneBattleService *services.ToneBattleService,
	userRepo *repositories.UserRepository,
	jwtSecret string,
) *ToneBattleHandler {
	return &ToneBattleHandler{
		toneBattleService: toneBattleService,
		userRepo:          userRepo,
		jwtSecret:         jwtSecret,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *ToneBattleHandler) ListQuestions(c *gin.Context) {
	limit := 50
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	questions, err := h.toneBattleService.ListQuestions(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tone battle questions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": questions,
		"total":     len(questions),
	})
}

func (h *ToneBattleHandler) HandleWebSocket(c *gin.Context) {
	user, err := h.authenticateBattleUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.toneBattleService.RegisterClient(user, conn)
	defer h.toneBattleService.UnregisterClient(user.ID)

	_ = conn.WriteJSON(gin.H{
		"type": "connected",
		"payload": gin.H{
			"userId":   user.ID,
			"nickname": user.Nickname,
		},
	})

	for {
		var message services.ToneBattleClientMessage
		if err := conn.ReadJSON(&message); err != nil {
			return
		}

		if err := h.toneBattleService.HandleClientMessage(user.ID, message); err != nil {
			_ = conn.WriteJSON(gin.H{
				"type": "error",
				"payload": gin.H{
					"message": err.Error(),
				},
			})
		}
	}
}

func (h *ToneBattleHandler) authenticateBattleUser(c *gin.Context) (*models.User, error) {
	tokenString := strings.TrimSpace(c.Query("token"))
	if tokenString == "" {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}
	if tokenString == "" {
		return nil, errors.New("Authorization token required")
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("Invalid or expired token")
	}

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("User not found")
	}

	return user, nil
}
