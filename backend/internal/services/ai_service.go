
package services

import (
	"chinese-learning-app/internal/utils"
)

type AIService struct {
	enabled bool
}

func NewAIService(apiKey string) *AIService {
	return &AIService{
		enabled: apiKey != "",
	}
}

// ========== 1. 语音转文字 ==========
func (s *AIService) SpeechToText(ctx interface{}, audioData []byte, filename string) (string, error) {
	// 简化版：返回模拟结果
	return "你好，我想学习中文", nil
}

// ========== 2. 发音评测 + 声调纠错 ==========
type PronunciationResult struct {
	Transcript string               `json:"transcript"`
	Score      int                  `json:"score"`
	Errors     []PronunciationError `json:"errors"`
	Feedback   string               `json:"feedback"`
}

type PronunciationError struct {
	Position   int    `json:"position"`
	Character  string `json:"character"`
	Expected   string `json:"expected"`
	Actual     string `json:"actual"`
	ErrorType  string `json:"errorType"`
}

func (s *AIService) EvaluatePronunciation(ctx interface{}, audioData []byte, expectedText string) (*PronunciationResult, error) {
	return s.mockEvaluatePronunciation(expectedText), nil
}

func (s *AIService) mockEvaluatePronunciation(expectedText string) *PronunciationResult {
	chars := utils.ExtractCharacters(expectedText)
	pinyin := utils.ToPinyin(expectedText)

	var mockErrors []PronunciationError
	if len(chars) > 1 {
		mockErrors = append(mockErrors, PronunciationError{
			Position:   0,
			Character:  chars[0],
			Expected:   pinyin[0],
			Actual:     pinyin[0] + " (声调模拟错误)",
			ErrorType:  "tone",
		})
	}

	score := 85
	if len(mockErrors) > 0 {
		score = 100 - len(mockErrors)*15
		if score < 0 {
			score = 0
		}
	}

	return &PronunciationResult{
		Transcript: expectedText,
		Score:      score,
		Errors:     mockErrors,
		Feedback:   "这是模拟模式下的评测结果。配置 OpenAI API Key 后可以获得真实评测！",
	}
}

// ========== 3. AI助教对话 ==========
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Reply       string              `json:"reply"`
	Corrections []Correction        `json:"corrections,omitempty"`
}

type Correction struct {
	Original   string `json:"original"`
	Suggestion string `json:"suggestion"`
	Reason     string `json:"reason"`
}

func (s *AIService) ChatWithTutor(ctx interface{}, userMessage string, scene string, history []ChatMessage) (*ChatResponse, error) {
	return s.mockChatWithTutor(userMessage, scene), nil
}

func (s *AIService) mockChatWithTutor(userMessage string, scene string) *ChatResponse {
	var reply string

	switch scene {
	case "restaurant":
		reply = "好的，我们有炒饭、面条和饺子。您想点什么？"
	case "airport":
		reply = "好的，您的航班在3号登机口，还有20分钟登机。请这边走！"
	case "hotel":
		reply = "好的，已为您预订了一个单人间，入住两晚。需要给您安排早餐吗？"
	case "shopping":
		reply = "这件T恤50元，买两件90元。您需要几件？"
	case "interview":
		reply = "请先简单介绍一下自己，然后我们再谈具体问题。"
	case "hospital":
		reply = "好的，请告诉我您的症状，我帮您安排相应的科室。"
	default:
		reply = "你好！很高兴和你练习中文。你今天有什么想聊的吗？"
	}

	return &ChatResponse{
		Reply:       reply + "（这是模拟模式，配置 OpenAI API Key 后可以获得真实AI回复！）",
		Corrections: []Correction{},
	}
}

func GetAvailableScenes() []string {
	return []string{"free_chat", "restaurant", "airport", "hotel", "shopping", "interview", "hospital"}
}
