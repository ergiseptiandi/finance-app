package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Service struct {
	apiKey      string
	apiURL      string
	model       string
	maxChats    int
	repo        Repository
	httpClient  *http.Client
}

func NewService(apiKey, apiURL, model string, maxChats int, repo Repository) *Service {
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/chat/completions"
	}
	if model == "" {
		model = "deepseek-chat"
	}
	return &Service{
		apiKey:     apiKey,
		apiURL:     apiURL,
		model:      model,
		maxChats:   maxChats,
		repo:       repo,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekRequest struct {
	Model       string            `json:"model"`
	Messages    []deepseekMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
}

type deepseekChoice struct {
	Message deepseekMessage `json:"message"`
}

type deepseekResponse struct {
	Choices []deepseekChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *Service) Analyze(ctx context.Context, userID int64, userName, message string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("AI tidak dikonfigurasi. Hubungi admin untuk mengatur DEEPSEEK_API_KEY.")
	}

	count, err := s.repo.GetChatCount(ctx, userID)
	if err != nil {
		return "", err
	}
	if count >= s.maxChats {
		return "", ErrChatLimitExceeded
	}

	systemPrompt := fmt.Sprintf(`Anda adalah asisten keuangan pribadi untuk aplikasi Finance-GO. 

DATA PENGGUNA:
- Nama: %s
- ID: %d

ATURAN:
1. Anda HANYA boleh menjawab pertanyaan seputar keuangan, analisis finansial, budgeting, utang, investasi, dan fitur-fitur di aplikasi Finance-GO.
2. JANGAN menjawab pertanyaan di luar topik keuangan (misal: "nama kamu siapa", "kamu bisa apa", dll). Beri tahu user bahwa Anda hanya fokus pada analisis keuangan.
3. Berikan jawaban yang singkat, jelas, dan actionable (bisa langsung dilakukan).
4. Gunakan bahasa Indonesia default, kecuali user bertanya dalam bahasa Inggris.
5. Jika user bertanya tentang data spesifik (saldo, transaksi, utang), beri tahu bahwa data real-time akan segera tersedia.
6. Rekomendasi harus spesifik dan praktis.
7. Jangan menyebut Anda adalah AI atau bot - cukup bantu saja.
8. Maksimal 3-4 kalimat per jawaban.`, userName, userID)

	promptTemplate := `Pertanyaan: %s

BANTUAN: Berikan analisis keuangan yang jelas dan actionable. Jika pertanyaan di luar keuangan, tolak dengan sopan dan arahkan kembali ke topik keuangan Finance-GO.`

	userPrompt := fmt.Sprintf(promptTemplate, message)

	reqBody := deepseekRequest{
		Model: s.model,
		Messages: []deepseekMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[ai] HTTP request failed: %v", err)
		return "", fmt.Errorf("AI API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ai] failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[ai] DeepSeek response status=%d, body=%s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var deepseekResp deepseekResponse
	if err := json.Unmarshal(respBody, &deepseekResp); err != nil {
		return "", fmt.Errorf("failed to parse AI response: %w", err)
	}

	if deepseekResp.Error != nil {
		return "", fmt.Errorf("AI API error: %s", deepseekResp.Error.Message)
	}

	if len(deepseekResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	reply := strings.TrimSpace(deepseekResp.Choices[0].Message.Content)

	if err := s.repo.IncrementChatCount(ctx, userID); err != nil {
		return "", err
	}

	return reply, nil
}

func (s *Service) GetUsage(ctx context.Context, userID int64) (UsageInfo, error) {
	count, err := s.repo.GetChatCount(ctx, userID)
	if err != nil {
		return UsageInfo{}, err
	}
	return UsageInfo{
		ChatCount: count,
		MaxChats:  s.maxChats,
	}, nil
}
