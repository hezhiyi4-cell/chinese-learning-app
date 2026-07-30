package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"chinese-learning-app/internal/config"
	"chinese-learning-app/internal/models"

	"github.com/gorilla/websocket"
)

const (
	xfyunTTSHost              = "tts-api.xfyun.cn"
	xfyunTTSPath              = "/v2/tts"
	pronunciationCacheVersion = "xfyun-v2"
	defaultPronunciationPause = `<break time="620ms"/>`
	shortPronunciationPause   = `<break time="420ms"/>`
)

var (
	whitespaceRE    = regexp.MustCompile(`\s+`)
	pinyinTokenRE   = regexp.MustCompile(`^[a-zv]+[1-5]?$`)
	nonTokenPunctRE = regexp.MustCompile(`[，,、；;。.!！？?／/\s]+`)

	initialOrder         = []string{"b", "p", "m", "f", "d", "t", "n", "l", "g", "k", "h", "j", "q", "x", "zh", "ch", "sh", "r", "z", "c", "s"}
	initialPronunciation = map[string]string{
		"b":  "波",
		"p":  "泼",
		"m":  "摸",
		"f":  "佛",
		"d":  "得",
		"t":  "特",
		"n":  "呢",
		"l":  "勒",
		"g":  "哥",
		"k":  "科",
		"h":  "喝",
		"j":  "鸡",
		"q":  "七",
		"x":  "西",
		"zh": "知",
		"ch": "吃",
		"sh": "诗",
		"r":  "日",
		"z":  "资",
		"c":  "雌",
		"s":  "思",
	}

	finalOrder         = []string{"a", "o", "e", "i", "u", "ü", "ai", "ei", "ui", "ao", "ou", "iu", "ie", "üe", "er", "an", "en", "in", "un", "ün", "ang", "eng", "ing", "ong"}
	finalPronunciation = map[string]string{
		"a":   "阿",
		"o":   "哦",
		"e":   "鹅",
		"i":   "衣",
		"u":   "乌",
		"ü":   "迂",
		"ai":  "哀",
		"ei":  "欸",
		"ui":  "威",
		"ao":  "袄",
		"ou":  "欧",
		"iu":  "优",
		"ie":  "耶",
		"üe":  "约",
		"er":  "儿",
		"an":  "安",
		"en":  "恩",
		"in":  "因",
		"un":  "温",
		"ün":  "晕",
		"ang": "昂",
		"eng": "亨",
		"ing": "英",
		"ong": "翁",
	}

	toneMarkedVowels = map[rune]struct {
		base string
		tone string
	}{
		'ā': {base: "a", tone: "1"}, 'á': {base: "a", tone: "2"}, 'ǎ': {base: "a", tone: "3"}, 'à': {base: "a", tone: "4"},
		'ē': {base: "e", tone: "1"}, 'é': {base: "e", tone: "2"}, 'ě': {base: "e", tone: "3"}, 'è': {base: "e", tone: "4"},
		'ī': {base: "i", tone: "1"}, 'í': {base: "i", tone: "2"}, 'ǐ': {base: "i", tone: "3"}, 'ì': {base: "i", tone: "4"},
		'ō': {base: "o", tone: "1"}, 'ó': {base: "o", tone: "2"}, 'ǒ': {base: "o", tone: "3"}, 'ò': {base: "o", tone: "4"},
		'ū': {base: "u", tone: "1"}, 'ú': {base: "u", tone: "2"}, 'ǔ': {base: "u", tone: "3"}, 'ù': {base: "u", tone: "4"},
		'ǖ': {base: "v", tone: "1"}, 'ǘ': {base: "v", tone: "2"}, 'ǚ': {base: "v", tone: "3"}, 'ǜ': {base: "v", tone: "4"},
		'ü': {base: "v", tone: ""},
	}
)

type PronunciationService struct {
	appID     string
	apiKey    string
	apiSecret string
	voice     string
	uploadDir string
}

type LessonPronunciationResult struct {
	Text     string
	AudioURL string
	Source   string
	Mode     string
}

type pronunciationScript struct {
	displayText string
	spokenText  string
}

type xfyunTTSFrame struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
	Data    struct {
		Audio  string `json:"audio"`
		Status int    `json:"status"`
	} `json:"data"`
}

func NewPronunciationService(cfg *config.Config) *PronunciationService {
	if cfg == nil {
		return &PronunciationService{}
	}

	return &PronunciationService{
		appID:     strings.TrimSpace(cfg.XFYunTTSAppID),
		apiKey:    strings.TrimSpace(cfg.XFYunTTSAPIKey),
		apiSecret: strings.TrimSpace(cfg.XFYunTTSAPISecret),
		voice:     strings.TrimSpace(cfg.XFYunTTSVoice),
		uploadDir: strings.TrimSpace(cfg.UploadDir),
	}
}

func (s *PronunciationService) GetLessonPronunciation(lesson *models.Lesson) (*LessonPronunciationResult, error) {
	if lesson == nil {
		return nil, errors.New("lesson is nil")
	}

	text := extractPronunciationSourceText(lesson)
	audioURL := strings.TrimSpace(lesson.AudioURL)
	if audioURL != "" {
		return &LessonPronunciationResult{
			Text:     text,
			AudioURL: audioURL,
			Source:   "lesson_audio",
			Mode:     "audio",
		}, nil
	}

	script := buildPronunciationScript(lesson.Title, text)
	if !s.Enabled() || strings.TrimSpace(script.spokenText) == "" {
		return &LessonPronunciationResult{
			Text:   script.displayText,
			Source: "xfyun_not_configured",
			Mode:   "text",
		}, nil
	}

	audioPath, relativeURL, err := s.ensureSynthesis(lesson, script)
	if err != nil {
		return &LessonPronunciationResult{
			Text:   script.displayText,
			Source: "xfyun_tts_unavailable",
			Mode:   "text",
		}, nil
	}

	_ = audioPath
	return &LessonPronunciationResult{
		Text:     script.displayText,
		AudioURL: relativeURL,
		Source:   "xfyun_tts",
		Mode:     "audio",
	}, nil
}

func (s *PronunciationService) Enabled() bool {
	return s.appID != "" && s.apiKey != "" && s.apiSecret != ""
}

func (s *PronunciationService) ensureSynthesis(lesson *models.Lesson, script pronunciationScript) (string, string, error) {
	cacheHash := buildPronunciationCacheKey(lesson, script.spokenText)
	baseName := fmt.Sprintf("lesson_%d_%s.mp3", lesson.ID, cacheHash[:12])
	targetDir := filepath.Join(s.resolvedUploadDir(), "tts")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", "", err
	}

	filePath := filepath.Join(targetDir, baseName)
	relativeURL := "/uploads/tts/" + baseName
	if _, err := os.Stat(filePath); err == nil {
		return filePath, relativeURL, nil
	}

	audioData, err := s.synthesize(script.spokenText)
	if err != nil {
		return "", "", err
	}
	if len(audioData) == 0 {
		return "", "", errors.New("empty pronunciation audio")
	}
	if err := os.WriteFile(filePath, audioData, 0o644); err != nil {
		return "", "", err
	}

	return filePath, relativeURL, nil
}

func (s *PronunciationService) synthesize(spokenText string) ([]byte, error) {
	callURL, err := s.buildAuthorizedURL()
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(callURL, nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	deadline := time.Now().Add(25 * time.Second)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, err
	}
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return nil, err
	}

	payload := map[string]any{
		"common": map[string]any{
			"app_id": s.appID,
		},
		"business": map[string]any{
			"aue":    "lame",
			"sfl":    1,
			"tte":    "utf8",
			"ttp":    "cssml",
			"vcn":    s.resolvedVoice(),
			"speed":  30,
			"pitch":  50,
			"volume": 70,
		},
		"data": map[string]any{
			"status":   2,
			"text":     []byte(spokenText),
			"encoding": "",
		},
	}

	if err := conn.WriteJSON(payload); err != nil {
		return nil, err
	}

	var audio []byte
	for {
		messageType, raw, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}
		if messageType != websocket.TextMessage {
			continue
		}

		var frame xfyunTTSFrame
		if err := json.Unmarshal(raw, &frame); err != nil {
			return nil, err
		}
		if frame.Code != 0 {
			return nil, fmt.Errorf("xfyun tts error %d: %s", frame.Code, frame.Message)
		}
		if strings.TrimSpace(frame.Data.Audio) != "" {
			chunk, err := base64.StdEncoding.DecodeString(frame.Data.Audio)
			if err != nil {
				return nil, err
			}
			audio = append(audio, chunk...)
		}
		if frame.Data.Status == 2 {
			return audio, nil
		}
	}
}

func (s *PronunciationService) buildAuthorizedURL() (string, error) {
	baseWSURL := "wss://" + xfyunTTSHost + xfyunTTSPath
	parsedURL, err := url.Parse(baseWSURL)
	if err != nil {
		return "", err
	}

	date := time.Now().UTC().Format(time.RFC1123)
	signatureOrigin := strings.Join([]string{
		"host: " + parsedURL.Host,
		"date: " + date,
		"GET " + parsedURL.Path + " HTTP/1.1",
	}, "\n")

	mac := hmac.New(sha256.New, []byte(s.apiSecret))
	_, _ = mac.Write([]byte(signatureOrigin))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	authorizationOrigin := fmt.Sprintf(
		`api_key="%s", algorithm="hmac-sha256", headers="host date request-line", signature="%s"`,
		s.apiKey,
		signature,
	)

	query := url.Values{}
	query.Set("host", parsedURL.Host)
	query.Set("date", date)
	query.Set("authorization", base64.StdEncoding.EncodeToString([]byte(authorizationOrigin)))

	return baseWSURL + "?" + query.Encode(), nil
}

func (s *PronunciationService) resolvedVoice() string {
	if s.voice != "" {
		return s.voice
	}
	return "xiaoyan"
}

func (s *PronunciationService) resolvedUploadDir() string {
	if s.uploadDir != "" {
		return s.uploadDir
	}
	return filepath.Join(".", "uploads")
}

func buildPronunciationCacheKey(lesson *models.Lesson, spokenText string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		pronunciationCacheVersion,
		fmt.Sprintf("%d", lesson.ID),
		lesson.UpdatedAt.UTC().Format(time.RFC3339Nano),
		strings.TrimSpace(lesson.Title),
		strings.TrimSpace(spokenText),
	}, "\n")))
	return hex.EncodeToString(sum[:])
}

func extractPronunciationSourceText(lesson *models.Lesson) string {
	if lesson == nil {
		return ""
	}

	raw := strings.TrimSpace(lesson.Content)
	if raw == "" {
		return strings.TrimSpace(lesson.Title)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err == nil {
		candidates := []string{"content", "text", "sentence", "chinese", "introduction"}
		for _, key := range candidates {
			value, _ := payload[key].(string)
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}

	return raw
}

func buildPronunciationScript(title, text string) pronunciationScript {
	displayText := strings.TrimSpace(text)
	if displayText == "" {
		displayText = strings.TrimSpace(title)
	}

	switch {
	case matchesPronunciationTable(displayText, initialOrder):
		return pronunciationScript{
			displayText: displayText,
			spokenText:  joinWithBreaks(initialOrder, initialPronunciation),
		}
	case matchesPronunciationTable(displayText, finalOrder):
		return pronunciationScript{
			displayText: displayText,
			spokenText:  joinWithBreaks(finalOrder, finalPronunciation),
		}
	case allTokensHaveTeachingMapping(displayText):
		return pronunciationScript{
			displayText: displayText,
			spokenText:  buildMappedTeachingCSSML(displayText),
		}
	case looksLikePinyinPhrase(displayText):
		return pronunciationScript{
			displayText: displayText,
			spokenText:  buildPinyinCSSML(displayText),
		}
	default:
		return pronunciationScript{
			displayText: displayText,
			spokenText:  escapeForCSSML(displayText),
		}
	}
}

func matchesPronunciationTable(text string, expected []string) bool {
	tokens := splitPronunciationTokens(text)
	if len(tokens) != len(expected) {
		return false
	}
	for index, token := range tokens {
		if token != expected[index] {
			return false
		}
	}
	return true
}

func splitPronunciationTokens(text string) []string {
	parts := nonTokenPunctRE.Split(strings.ToLower(strings.TrimSpace(text)), -1)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func joinWithBreaks(order []string, dictionary map[string]string) string {
	parts := make([]string, 0, len(order))
	for index, token := range order {
		spoken := strings.TrimSpace(dictionary[token])
		if spoken == "" {
			spoken = token
		}
		parts = append(parts, escapeForCSSML(spoken))
		if index < len(order)-1 {
			parts = append(parts, defaultPronunciationPause)
		}
	}
	return strings.Join(parts, "")
}

func allTokensHaveTeachingMapping(text string) bool {
	tokens := splitPronunciationTokens(text)
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if _, ok := lookupTeachingPronunciation(token); !ok {
			return false
		}
	}
	return true
}

func buildMappedTeachingCSSML(text string) string {
	tokens := splitPronunciationTokens(text)
	parts := make([]string, 0, len(tokens)*2)
	for index, token := range tokens {
		spoken, ok := lookupTeachingPronunciation(token)
		if !ok {
			spoken = token
		}
		parts = append(parts, escapeForCSSML(spoken))
		if index < len(tokens)-1 {
			parts = append(parts, defaultPronunciationPause)
		}
	}
	return strings.Join(parts, "")
}

func lookupTeachingPronunciation(token string) (string, bool) {
	normalized := normalizeTeachingToken(token)
	if normalized == "" {
		return "", false
	}
	if spoken, ok := initialPronunciation[normalized]; ok {
		return spoken, true
	}
	if spoken, ok := finalPronunciation[normalized]; ok {
		return spoken, true
	}
	return "", false
}

func normalizeTeachingToken(token string) string {
	normalized := strings.ToLower(strings.TrimSpace(token))
	normalized = strings.ReplaceAll(normalized, "u:", "ü")
	normalized = strings.ReplaceAll(normalized, "v", "ü")
	return normalized
}

func looksLikePinyinPhrase(text string) bool {
	tokens := splitPronunciationTokens(text)
	if len(tokens) == 0 {
		return false
	}
	for _, token := range tokens {
		if _, ok := normalizePinyinToken(token); !ok {
			return false
		}
	}
	return true
}

func buildPinyinCSSML(text string) string {
	tokens := splitPronunciationTokens(text)
	parts := make([]string, 0, len(tokens)*2)
	for index, token := range tokens {
		normalized, ok := normalizePinyinToken(token)
		if !ok {
			parts = append(parts, escapeForCSSML(token))
		} else {
			parts = append(parts, `<phoneme lang="zh-cn">`+normalized+`</phoneme>`)
		}
		if index < len(tokens)-1 {
			parts = append(parts, shortPronunciationPause)
		}
	}
	return strings.Join(parts, "")
}

func normalizePinyinToken(token string) (string, bool) {
	raw := strings.ToLower(strings.TrimSpace(token))
	if raw == "" {
		return "", false
	}

	raw = strings.ReplaceAll(raw, "u:", "v")
	raw = strings.ReplaceAll(raw, "ü", "v")
	raw = strings.ReplaceAll(raw, "’", "")
	raw = strings.ReplaceAll(raw, "'", "")

	var builder strings.Builder
	tone := ""
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '1' && r <= '5':
			tone = string(r)
		default:
			if marked, ok := toneMarkedVowels[r]; ok {
				builder.WriteString(marked.base)
				if marked.tone != "" {
					tone = marked.tone
				}
				continue
			}
			if unicode.IsSpace(r) {
				continue
			}
			return "", false
		}
	}

	base := builder.String()
	if base == "" || !pinyinTokenRE.MatchString(base) {
		return "", false
	}
	if tone != "" {
		return base + tone, true
	}
	return base, true
}

func escapeForCSSML(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return replacer.Replace(strings.TrimSpace(whitespaceRE.ReplaceAllString(text, " ")))
}
