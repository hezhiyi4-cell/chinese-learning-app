package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"chinese-learning-app/internal/models"
	"chinese-learning-app/internal/repositories"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	ToneBattleTargetScore = 10
	ToneBattleWinXP       = 20
	ToneBattleLoseXP      = 5
)

var toneBattleBaseSyllablePattern = regexp.MustCompile(`([a-züv]+)`)

type ToneBattleClient struct {
	UserID   uint
	Nickname string
	Conn     *websocket.Conn
	RoomID   string
	Matching bool
	mu       sync.Mutex
}

type ToneBattlePlayerState struct {
	UserID       uint
	Nickname     string
	Score        int
	SelectedTone int
	Answered     bool
	Correct      bool
}

type ToneBattleRoom struct {
	ID               string
	Match            *models.ToneBattleMatch
	PlayerOne        *ToneBattlePlayerState
	PlayerTwo        *ToneBattlePlayerState
	CurrentQuestion  *models.ToneBattleQuestion
	AskedQuestionIDs []uint
	Round            int
	Finished         bool
}

type ToneBattleQuestionPrompt struct {
	ID          uint   `json:"id"`
	DisplayText string `json:"displayText"`
	Hanzi       string `json:"hanzi"`
	Pinyin      string `json:"pinyin"`
	AudioPath   string `json:"audioPath"`
}

type ToneBattleClientMessage struct {
	Type string `json:"type"`
	Tone int    `json:"tone,omitempty"`
}

type ToneBattleQuestionSummary struct {
	ID        uint   `json:"id"`
	Hanzi     string `json:"hanzi"`
	Syllable  string `json:"syllable"`
	Pinyin    string `json:"pinyin"`
	Tone      int    `json:"tone"`
	AudioPath string `json:"audioPath"`
}

type ToneBattleService struct {
	questionRepo *repositories.ToneBattleRepository
	userRepo     *repositories.UserRepository
	redisClient  *redis.Client

	mu           sync.Mutex
	clients      map[uint]*ToneBattleClient
	waitingQueue []uint
	waitingSet   map[uint]bool
	rooms        map[string]*ToneBattleRoom
}

func NewToneBattleService(
	questionRepo *repositories.ToneBattleRepository,
	userRepo *repositories.UserRepository,
	redisAddr string,
	redisPassword string,
) *ToneBattleService {
	service := &ToneBattleService{
		questionRepo: questionRepo,
		userRepo:     userRepo,
		clients:      make(map[uint]*ToneBattleClient),
		waitingSet:   make(map[uint]bool),
		rooms:        make(map[string]*ToneBattleRoom),
	}

	if redisAddr != "" {
		client := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			log.Printf("ToneBattle: Redis unavailable, use in-memory queue instead: %v", err)
		} else {
			service.redisClient = client
		}
	}

	return service
}

func (s *ToneBattleService) RegisterClient(user *models.User, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[user.ID] = &ToneBattleClient{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Conn:     conn,
	}
}

func (s *ToneBattleService) UnregisterClient(userID uint) {
	var opponent *ToneBattleClient
	var gameOverPayload map[string]any

	s.mu.Lock()
	client := s.clients[userID]
	if client == nil {
		s.mu.Unlock()
		return
	}

	s.removeFromQueueLocked(userID)
	if client.RoomID != "" {
		room := s.rooms[client.RoomID]
		if room != nil && !room.Finished {
			room.Finished = true
			winner, loser := s.resolveForfeitLocked(room, userID)
			delete(s.rooms, room.ID)
			if winner != nil {
				opponent = s.clients[winner.UserID]
				gameOverPayload = s.buildGameOverPayload(room, winner.UserID, ToneBattleWinXP, 0, true)
				client.RoomID = ""
				if loserClient := s.clients[loser.UserID]; loserClient != nil {
					loserClient.RoomID = ""
				}
			}
		}
	}
	delete(s.clients, userID)
	s.persistWaitingPoolLocked()
	s.mu.Unlock()

	if opponent != nil && gameOverPayload != nil {
		_ = s.userRepo.AddXP(opponent.UserID, ToneBattleWinXP)
		_ = s.sendToClient(opponent, "game_over", gameOverPayload)
	}
}

func (s *ToneBattleService) HandleClientMessage(userID uint, message ToneBattleClientMessage) error {
	switch message.Type {
	case "start_match":
		return s.startMatchmaking(userID)
	case "cancel_match":
		return s.cancelMatchmaking(userID)
	case "submit_answer":
		return s.submitAnswer(userID, message.Tone)
	case "leave_room":
		return s.leaveRoom(userID)
	default:
		return errors.New("unsupported tone battle message type")
	}
}

func (s *ToneBattleService) ListQuestions(limit int) ([]ToneBattleQuestionSummary, error) {
	questions, err := s.questionRepo.ListQuestions(limit)
	if err != nil {
		return nil, err
	}

	result := make([]ToneBattleQuestionSummary, 0, len(questions))
	for _, question := range questions {
		result = append(result, ToneBattleQuestionSummary{
			ID:        question.ID,
			Hanzi:     question.Hanzi,
			Syllable:  question.Syllable,
			Pinyin:    question.Pinyin,
			Tone:      question.Tone,
			AudioPath: question.AudioPath,
		})
	}
	return result, nil
}

func (s *ToneBattleService) startMatchmaking(userID uint) error {
	var pair [2]uint
	var shouldStartRoom bool

	s.mu.Lock()
	client := s.clients[userID]
	if client == nil {
		s.mu.Unlock()
		return errors.New("connection not registered")
	}
	if client.RoomID != "" {
		s.mu.Unlock()
		return errors.New("user is already in a battle room")
	}
	if !s.waitingSet[userID] {
		s.waitingQueue = append(s.waitingQueue, userID)
		s.waitingSet[userID] = true
		client.Matching = true
	}
	s.persistWaitingPoolLocked()

	first, second, ok := s.popPairLocked()
	if ok {
		pair = [2]uint{first, second}
		shouldStartRoom = true
	}
	s.mu.Unlock()

	if shouldStartRoom {
		return s.startRoom(pair[0], pair[1])
	}

	return s.notifyUserStatus(userID, "matching", map[string]any{
		"message": "正在匹配对手...",
	})
}

func (s *ToneBattleService) cancelMatchmaking(userID uint) error {
	s.mu.Lock()
	s.removeFromQueueLocked(userID)
	s.persistWaitingPoolLocked()
	client := s.clients[userID]
	s.mu.Unlock()

	if client == nil {
		return nil
	}
	return s.sendToClient(client, "match_cancelled", map[string]any{
		"message": "已取消匹配",
	})
}

func (s *ToneBattleService) leaveRoom(userID uint) error {
	s.mu.Lock()
	client := s.clients[userID]
	if client == nil || client.RoomID == "" {
		s.mu.Unlock()
		return nil
	}
	room := s.rooms[client.RoomID]
	if room == nil {
		client.RoomID = ""
		s.mu.Unlock()
		return nil
	}

	room.Finished = true
	winner, loser := s.resolveForfeitLocked(room, userID)
	delete(s.rooms, room.ID)
	if winner != nil {
		if winnerClient := s.clients[winner.UserID]; winnerClient != nil {
			winnerClient.RoomID = ""
		}
		if loserClient := s.clients[loser.UserID]; loserClient != nil {
			loserClient.RoomID = ""
		}
	}
	s.mu.Unlock()

	if winner == nil {
		return nil
	}

	_ = s.userRepo.AddXP(winner.UserID, ToneBattleWinXP)
	if winnerClient := s.getClient(winner.UserID); winnerClient != nil {
		return s.sendToClient(winnerClient, "game_over", s.buildGameOverPayload(room, winner.UserID, ToneBattleWinXP, 0, true))
	}
	return nil
}

func (s *ToneBattleService) submitAnswer(userID uint, tone int) error {
	if tone < 1 || tone > 5 {
		return errors.New("tone must be between 1 and 5")
	}

	var roomID string
	var payload map[string]any
	var shouldFinish bool
	var winnerID uint
	var winnerReward int
	var loserReward int

	s.mu.Lock()
	client := s.clients[userID]
	if client == nil || client.RoomID == "" {
		s.mu.Unlock()
		return errors.New("user is not in an active room")
	}
	room := s.rooms[client.RoomID]
	if room == nil || room.CurrentQuestion == nil || room.Finished {
		s.mu.Unlock()
		return errors.New("battle room is not ready")
	}

	player := room.playerByUserID(userID)
	if player == nil {
		s.mu.Unlock()
		return errors.New("player not found")
	}
	if player.Answered {
		s.mu.Unlock()
		return errors.New("answer already submitted")
	}

	player.Answered = true
	player.SelectedTone = tone
	player.Correct = tone == room.CurrentQuestion.Tone
	if player.Correct {
		player.Score++
	} else {
		player.Score--
	}

	roomID = room.ID
	payload = s.buildRoundPayload(room)
	shouldFinish, winnerID = s.resolveWinnerLocked(room)
	if shouldFinish {
		room.Finished = true
		winnerReward = ToneBattleWinXP
		loserReward = ToneBattleLoseXP
		delete(s.rooms, room.ID)
		if c := s.clients[room.PlayerOne.UserID]; c != nil {
			c.RoomID = ""
		}
		if c := s.clients[room.PlayerTwo.UserID]; c != nil {
			c.RoomID = ""
		}
	}
	s.mu.Unlock()

	if err := s.broadcastRoom(roomID, "round_result", payload); err != nil {
		log.Printf("ToneBattle: broadcast round result failed: %v", err)
	}

	if shouldFinish {
		var loserID uint
		if room.PlayerOne.UserID == winnerID {
			loserID = room.PlayerTwo.UserID
		} else {
			loserID = room.PlayerOne.UserID
		}
		_ = s.userRepo.AddXP(winnerID, winnerReward)
		_ = s.userRepo.AddXP(loserID, loserReward)
		if winnerClient := s.getClient(winnerID); winnerClient != nil {
			_ = s.sendToClient(winnerClient, "game_over", s.buildGameOverPayload(room, winnerID, winnerReward, loserReward, false))
		}
		if loserClient := s.getClient(loserID); loserClient != nil {
			_ = s.sendToClient(loserClient, "game_over", s.buildGameOverPayload(room, winnerID, winnerReward, loserReward, false))
		}
		return nil
	}

	go func() {
		time.Sleep(1500 * time.Millisecond)
		if err := s.pushNextQuestion(roomID); err != nil {
			log.Printf("ToneBattle: push next question failed: %v", err)
		}
	}()

	return nil
}

func (s *ToneBattleService) startRoom(firstUserID, secondUserID uint) error {
	s.mu.Lock()
	firstClient := s.clients[firstUserID]
	secondClient := s.clients[secondUserID]
	s.mu.Unlock()

	if firstClient == nil || secondClient == nil {
		return errors.New("matched user disconnected before room creation")
	}

	now := time.Now()
	roomID := fmt.Sprintf("tone-room-%d-%d-%d", firstUserID, secondUserID, now.UnixNano())
	match := &models.ToneBattleMatch{
		RoomID:      roomID,
		PlayerOneID: firstUserID,
		PlayerTwoID: secondUserID,
		Status:      "active",
		StartedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.questionRepo.CreateMatch(match); err != nil {
		return err
	}

	room := &ToneBattleRoom{
		ID:    roomID,
		Match: match,
		PlayerOne: &ToneBattlePlayerState{
			UserID:   firstUserID,
			Nickname: fallbackNickname(firstClient.Nickname, firstUserID),
		},
		PlayerTwo: &ToneBattlePlayerState{
			UserID:   secondUserID,
			Nickname: fallbackNickname(secondClient.Nickname, secondUserID),
		},
	}

	s.mu.Lock()
	s.rooms[roomID] = room
	firstClient.RoomID = roomID
	firstClient.Matching = false
	secondClient.RoomID = roomID
	secondClient.Matching = false
	s.mu.Unlock()

	_ = s.sendToClient(firstClient, "matched", map[string]any{
		"roomId":      roomID,
		"self":        room.playerSnapshot(firstUserID),
		"opponent":    room.playerSnapshot(secondUserID),
		"countdown":   3,
		"targetScore": ToneBattleTargetScore,
	})
	_ = s.sendToClient(secondClient, "matched", map[string]any{
		"roomId":      roomID,
		"self":        room.playerSnapshot(secondUserID),
		"opponent":    room.playerSnapshot(firstUserID),
		"countdown":   3,
		"targetScore": ToneBattleTargetScore,
	})

	go func() {
		for countdown := 3; countdown >= 1; countdown-- {
			_ = s.broadcastRoom(roomID, "countdown", map[string]any{
				"seconds": countdown,
			})
			time.Sleep(time.Second)
		}
		if err := s.pushNextQuestion(roomID); err != nil {
			log.Printf("ToneBattle: initial question failed: %v", err)
		}
	}()

	return nil
}

func (s *ToneBattleService) pushNextQuestion(roomID string) error {
	s.mu.Lock()
	room := s.rooms[roomID]
	if room == nil || room.Finished {
		s.mu.Unlock()
		return nil
	}
	excludeIDs := append([]uint(nil), room.AskedQuestionIDs...)
	s.mu.Unlock()

	question, err := s.questionRepo.RandomQuestion(excludeIDs)
	if err != nil {
		return err
	}
	if question == nil {
		question, err = s.questionRepo.RandomQuestion(nil)
		if err != nil {
			return err
		}
		if question == nil {
			return errors.New("tone battle question bank is empty")
		}
	}

	s.mu.Lock()
	room = s.rooms[roomID]
	if room == nil || room.Finished {
		s.mu.Unlock()
		return nil
	}
	room.CurrentQuestion = question
	room.AskedQuestionIDs = append(room.AskedQuestionIDs, question.ID)
	room.Round++
	room.PlayerOne.Answered = false
	room.PlayerOne.SelectedTone = 0
	room.PlayerOne.Correct = false
	room.PlayerTwo.Answered = false
	room.PlayerTwo.SelectedTone = 0
	room.PlayerTwo.Correct = false
	s.mu.Unlock()

	return s.broadcastRoom(roomID, "question", map[string]any{
		"round": room.Round,
		"question": ToneBattleQuestionPrompt{
			ID:          question.ID,
			DisplayText: baseSyllable(question.Syllable),
			Hanzi:       question.Hanzi,
			Pinyin:      question.Pinyin,
			AudioPath:   question.AudioPath,
		},
	})
}

func (s *ToneBattleService) buildRoundPayload(room *ToneBattleRoom) map[string]any {
	return map[string]any{
		"questionId":    room.CurrentQuestion.ID,
		"correctTone":   room.CurrentQuestion.Tone,
		"correctPinyin": room.CurrentQuestion.Pinyin,
		"correctHanzi":  room.CurrentQuestion.Hanzi,
		"players": []map[string]any{
			{
				"userId":       room.PlayerOne.UserID,
				"nickname":     room.PlayerOne.Nickname,
				"selectedTone": room.PlayerOne.SelectedTone,
				"correct":      room.PlayerOne.Correct,
				"score":        room.PlayerOne.Score,
			},
			{
				"userId":       room.PlayerTwo.UserID,
				"nickname":     room.PlayerTwo.Nickname,
				"selectedTone": room.PlayerTwo.SelectedTone,
				"correct":      room.PlayerTwo.Correct,
				"score":        room.PlayerTwo.Score,
			},
		},
	}
}

func (s *ToneBattleService) buildGameOverPayload(
	room *ToneBattleRoom,
	winnerID uint,
	winnerReward int,
	loserReward int,
	byForfeit bool,
) map[string]any {
	return map[string]any{
		"roomId":   room.ID,
		"winnerId": winnerID,
		"playerOne": map[string]any{
			"userId":   room.PlayerOne.UserID,
			"nickname": room.PlayerOne.Nickname,
			"score":    room.PlayerOne.Score,
		},
		"playerTwo": map[string]any{
			"userId":   room.PlayerTwo.UserID,
			"nickname": room.PlayerTwo.Nickname,
			"score":    room.PlayerTwo.Score,
		},
		"winnerRewardXp": winnerReward,
		"loserRewardXp":  loserReward,
		"byForfeit":      byForfeit,
	}
}

func (s *ToneBattleService) resolveWinnerLocked(room *ToneBattleRoom) (bool, uint) {
	if !room.PlayerOne.Answered || !room.PlayerTwo.Answered {
		return false, 0
	}

	playerOneReached := room.PlayerOne.Score >= ToneBattleTargetScore
	playerTwoReached := room.PlayerTwo.Score >= ToneBattleTargetScore

	switch {
	case playerOneReached && playerTwoReached:
		if room.PlayerOne.Score > room.PlayerTwo.Score {
			return true, room.PlayerOne.UserID
		}
		if room.PlayerTwo.Score > room.PlayerOne.Score {
			return true, room.PlayerTwo.UserID
		}
		return false, 0
	case playerOneReached:
		return true, room.PlayerOne.UserID
	case playerTwoReached:
		return true, room.PlayerTwo.UserID
	default:
		return false, 0
	}
}

func (s *ToneBattleService) resolveForfeitLocked(room *ToneBattleRoom, quitterID uint) (*ToneBattlePlayerState, *ToneBattlePlayerState) {
	if room.PlayerOne.UserID == quitterID {
		room.PlayerTwo.Score = max(room.PlayerTwo.Score, ToneBattleTargetScore)
		return room.PlayerTwo, room.PlayerOne
	}
	room.PlayerOne.Score = max(room.PlayerOne.Score, ToneBattleTargetScore)
	return room.PlayerOne, room.PlayerTwo
}

func (s *ToneBattleService) broadcastRoom(roomID, eventType string, payload map[string]any) error {
	s.mu.Lock()
	room := s.rooms[roomID]
	if room == nil && eventType != "game_over" {
		s.mu.Unlock()
		return nil
	}

	var recipients []*ToneBattleClient
	if room != nil {
		if client := s.clients[room.PlayerOne.UserID]; client != nil {
			recipients = append(recipients, client)
		}
		if client := s.clients[room.PlayerTwo.UserID]; client != nil {
			recipients = append(recipients, client)
		}
	}
	s.mu.Unlock()

	for _, client := range recipients {
		if err := s.sendToClient(client, eventType, payload); err != nil {
			log.Printf("ToneBattle: failed to send %s to user %d: %v", eventType, client.UserID, err)
		}
	}
	return nil
}

func (s *ToneBattleService) sendToClient(client *ToneBattleClient, eventType string, payload map[string]any) error {
	if client == nil || client.Conn == nil {
		return nil
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	return client.Conn.WriteJSON(map[string]any{
		"type":    eventType,
		"payload": payload,
	})
}

func (s *ToneBattleService) notifyUserStatus(userID uint, eventType string, payload map[string]any) error {
	client := s.getClient(userID)
	if client == nil {
		return nil
	}
	return s.sendToClient(client, eventType, payload)
}

func (s *ToneBattleService) getClient(userID uint) *ToneBattleClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clients[userID]
}

func (s *ToneBattleService) popPairLocked() (uint, uint, bool) {
	cleaned := make([]uint, 0, len(s.waitingQueue))
	for _, userID := range s.waitingQueue {
		client := s.clients[userID]
		if client == nil || client.RoomID != "" || !s.waitingSet[userID] {
			delete(s.waitingSet, userID)
			continue
		}
		cleaned = append(cleaned, userID)
	}
	s.waitingQueue = cleaned
	if len(s.waitingQueue) < 2 {
		return 0, 0, false
	}

	first := s.waitingQueue[0]
	second := s.waitingQueue[1]
	s.waitingQueue = s.waitingQueue[2:]
	delete(s.waitingSet, first)
	delete(s.waitingSet, second)
	return first, second, true
}

func (s *ToneBattleService) removeFromQueueLocked(userID uint) {
	delete(s.waitingSet, userID)
	filtered := make([]uint, 0, len(s.waitingQueue))
	for _, item := range s.waitingQueue {
		if item != userID {
			filtered = append(filtered, item)
		}
	}
	s.waitingQueue = filtered
	if client := s.clients[userID]; client != nil {
		client.Matching = false
	}
}

func (s *ToneBattleService) persistWaitingPoolLocked() {
	if s.redisClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "tone_battle:match_pool"
	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		return
	}
	if len(s.waitingQueue) == 0 {
		return
	}

	members := make([]any, 0, len(s.waitingQueue))
	for _, userID := range s.waitingQueue {
		members = append(members, strconv.FormatUint(uint64(userID), 10))
	}
	_ = s.redisClient.RPush(ctx, key, members...).Err()
}

func (r *ToneBattleRoom) playerByUserID(userID uint) *ToneBattlePlayerState {
	if r.PlayerOne.UserID == userID {
		return r.PlayerOne
	}
	if r.PlayerTwo.UserID == userID {
		return r.PlayerTwo
	}
	return nil
}

func (r *ToneBattleRoom) playerSnapshot(userID uint) map[string]any {
	player := r.playerByUserID(userID)
	if player == nil {
		return map[string]any{}
	}
	return map[string]any{
		"userId":   player.UserID,
		"nickname": player.Nickname,
		"score":    player.Score,
	}
}

func baseSyllable(syllable string) string {
	match := toneBattleBaseSyllablePattern.FindStringSubmatch(syllable)
	if len(match) > 1 {
		return match[1]
	}
	return syllable
}

func fallbackNickname(nickname string, userID uint) string {
	if nickname != "" {
		return nickname
	}
	return fmt.Sprintf("玩家%d", userID)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
