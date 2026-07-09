
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL    = "http://localhost:8080"
	testEmail  = "test_ai_user@example.com"
	testPass   = "123456"
	testNick   = "AI 测试用户"
)

type AuthResponse struct {
	Token string                 `json:"token"`
	User  map[string]interface{} `json:"user"`
}

type CourseListResponse struct {
	Courses []map[string]interface{} `json:"courses"`
}

type ChatRequest struct {
	Message string                   `json:"message"`
	Scene   string                   `json:"scene"`
	History []map[string]interface{} `json:"history"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func main() {
	fmt.Println("=======================================")
	fmt.Println("   中文学习App - AI 功能测试脚本")
	fmt.Println("=======================================")
	fmt.Println()

	fmt.Println("📋 测试说明:")
	fmt.Println("- 此脚本测试 API 流程")
	fmt.Println("- AI 功能当前处于模拟模式")
	fmt.Println("- 配置 OpenAI API Key 后可获取真实响应")
	fmt.Println()

	// 步骤 1: 注册用户
	fmt.Println("\n===== 步骤 1: 注册测试账号 =====")
	token := registerTestUser()
	if token == "" {
		fmt.Println("注册失败，尝试登录现有账号...")
		token = loginTestUser()
		if token == "" {
			fmt.Println("❌ 无法获取 Token，退出测试")
			return
		}
	}
	fmt.Println("✅ 登录成功! Token 已获取")

	// 步骤 2: 获取课程列表
	fmt.Println("\n===== 步骤 2: 获取课程列表 =====")
	testGetCourses()

	// 步骤 3: 获取 AI 场景列表
	fmt.Println("\n===== 步骤 3: 获取 AI 对话场景 =====")
	testGetAIScenes(token)

	// 步骤 4: 测试 AI 对话功能
	fmt.Println("\n===== 步骤 4: 测试 AI 助教对话 =====")
	testAIChat(token, "restaurant", "你好，我想点一份炒饭")
	testAIChat(token, "free_chat", "你好，我想练习中文")

	fmt.Println("\n=======================================")
	fmt.Println("   ✅ 所有测试完成！")
	fmt.Println("=======================================")
	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("- 如需测试真实 AI 功能，请配置 OPENAI_API_KEY")
	fmt.Println("- 音频测试需使用 Postman 或类似工具上传音频文件")
	fmt.Println()
}

func registerTestUser() string {
	fmt.Println("正在注册账号:", testEmail)
	reqBody := map[string]string{
		"email":     testEmail,
		"password":  testPass,
		"nickname":  testNick,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("注册请求失败:", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		fmt.Println("⚠️  账号已存在")
		return ""
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		fmt.Printf("注册失败，状态码: %d\n", resp.StatusCode)
		return ""
	}

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)
	fmt.Println("✅ 账号注册成功!")
	return authResp.Token
}

func loginTestUser() string {
	fmt.Println("正在登录账号:", testEmail)
	reqBody := map[string]string{
		"email":    testEmail,
		"password": testPass,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("登录请求失败:", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("登录失败，状态码: %d\n", resp.StatusCode)
		return ""
	}

	var authResp AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)
	return authResp.Token
}

func testGetCourses() {
	resp, err := http.Get(baseURL + "/api/v1/courses")
	if err != nil {
		fmt.Println("获取课程失败:", err)
		return
	}
	defer resp.Body.Close()

	var courseResp CourseListResponse
	json.NewDecoder(resp.Body).Decode(&courseResp)
	fmt.Printf("✅ 获取到 %d 个课程:\n", len(courseResp.Courses))
	for i, course := range courseResp.Courses {
		fmt.Printf("  %d. [%s] %s\n", i+1, course["level"], course["title"])
	}
}

func testGetAIScenes(token string) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/ai/scenes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("获取场景失败:", err)
		return
	}
	defer resp.Body.Close()

	var sceneResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sceneResp)

	if scenes, ok := sceneResp["scenes"].([]interface{}); ok {
		fmt.Printf("✅ 获取到 %d 个可用场景:\n", len(scenes))
		for _, s := range scenes {
			scene := s.(map[string]interface{})
			fmt.Printf("  - [%s] %s\n", scene["id"], scene["name"])
		}
	}
}

func testAIChat(token string, scene, message string) {
	fmt.Printf("\n🎯 测试场景: [%s] 用户消息: \"%s\"\n", scene, message)

	reqBody := ChatRequest{
		Message: message,
		Scene:   scene,
		History: []map[string]interface{}{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/ai/chat", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("AI 对话请求失败:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("AI 对话失败 (状态码: %d): %s\n", resp.StatusCode, respBody)
		return
	}

	var chatResp ChatResponse
	json.NewDecoder(resp.Body).Decode(&chatResp)
	fmt.Printf("🤖 AI 回复: \"%s\"\n", chatResp.Reply)
	fmt.Println("✅ 对话测试通过!")
}

// 测试音频上传（需准备真实音频文件）
func testAudioUpload(token string, expectedText string) {
	fmt.Println("\n🎤 音频评测测试（需要真实音频文件）")
	fmt.Println("   提示: 请使用 Postman 或 curl 上传音频")
	fmt.Println("   API: POST /api/v1/ai/evaluate")
	fmt.Println("   参数: audio（文件）, expectedText（文本）")
}
