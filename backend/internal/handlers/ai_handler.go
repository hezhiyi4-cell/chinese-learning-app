
package handlers

import (
	"chinese-learning-app/internal/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiService *services.AIService
}

func NewAIHandler(aiService *services.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// ========== 1. Speech to Text ==========
func (h *AIHandler) SpeechToText(c *gin.Context) {
	// 获取音频文件
	file, _, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}
	defer file.Close()

	// 读取文件内容
	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read audio file"})
		return
	}

	// 检查文件大小（10MB）
	if len(audioData) > 10*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "audio file too large (max 10MB)"})
		return
	}

	// 调用 AI Service
	text, err := h.aiService.SpeechToText(c.Request.Context(), audioData, "audio.wav")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"text": text,
	})
}

// ========== 2. Evaluate Pronunciation ==========
func (h *AIHandler) Evaluate(c *gin.Context) {
	// 获取预期文本
	expectedText := c.PostForm("expectedText")
	if expectedText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expectedText is required"})
		return
	}

	// 获取音频文件
	file, _, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}
	defer file.Close()

	// 读取文件内容
	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read audio file"})
		return
	}

	// 检查文件大小（10MB）
	if len(audioData) > 10*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "audio file too large (max 10MB)"})
		return
	}

	// 调用 AI Service
	result, err := h.aiService.EvaluatePronunciation(c.Request.Context(), audioData, expectedText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ========== 3. Chat With Tutor ==========
type ChatRequest struct {
	Message string                 `json:"message" binding:"required"`
	Scene   string                 `json:"scene"`
	History []services.ChatMessage `json:"history"`
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Scene == "" {
		req.Scene = "free_chat"
	}

	result, err := h.aiService.ChatWithTutor(c.Request.Context(), req.Message, req.Scene, req.History)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ========== 4. Get Scenes ==========
func (h *AIHandler) GetScenes(c *gin.Context) {
	scenes := []map[string]string{
		{"id": "free_chat", "name": "自由对话", "description": "随意聊天，练习口语"},
		{"id": "restaurant", "name": "餐厅点餐", "description": "在餐厅吃饭、点餐的场景"},
		{"id": "airport", "name": "机场场景", "description": "在机场乘机、问路的场景"},
		{"id": "hotel", "name": "酒店入住", "description": "在酒店入住、询问的场景"},
		{"id": "shopping", "name": "商店购物", "description": "在商店购物的场景"},
		{"id": "interview", "name": "面试场景", "description": "中文面试练习"},
		{"id": "hospital", "name": "医院看病", "description": "在医院看病的场景"},
	}

	c.JSON(http.StatusOK, gin.H{
		"scenes": scenes,
	})
}
