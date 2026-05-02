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

type FinancialDataProvider interface {
	GetFinancialSummary(ctx context.Context, userID int64) (FinancialSummary, error)
}

type FinancialSummary struct {
	TotalBalance       float64              `json:"total_balance"`
	MonthlyIncome      float64              `json:"monthly_income"`
	MonthlyExpense     float64              `json:"monthly_expense"`
	ConsumptionExpense float64              `json:"consumption_expense"`
	DebtRepayment      float64              `json:"debt_repayment"`
	NetCashflow        float64              `json:"net_cashflow"`
	SavingsRate        float64              `json:"savings_rate"`
	ExpenseRatio       float64              `json:"expense_ratio"`
	DebtTotal          float64              `json:"debt_total"`
	DebtRemaining      float64              `json:"debt_remaining"`
	DebtCompletionRate float64              `json:"debt_completion_rate"`
	BudgetUsage        float64              `json:"budget_usage"`
	BudgetRemaining    float64              `json:"budget_remaining"`
	CategoryBreakdown  []CategorySummary    `json:"category_breakdown,omitempty"`
	RecentTransactions int                  `json:"recent_transactions"`
}

type CategorySummary struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type Service struct {
	apiKey         string
	apiURL         string
	model          string
	maxChats       int
	repo           Repository
	dataProvider   FinancialDataProvider
	httpClient     *http.Client
}

func NewService(apiKey, apiURL, model string, maxChats int, repo Repository, dataProvider FinancialDataProvider) *Service {
	if apiURL == "" {
		apiURL = "https://api.deepseek.com/chat/completions"
	}
	if model == "" {
		model = "deepseek-chat"
	}
	return &Service{
		apiKey:       apiKey,
		apiURL:       apiURL,
		model:        model,
		maxChats:     maxChats,
		repo:         repo,
		dataProvider: dataProvider,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
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

	financialData := ""
	if s.dataProvider != nil {
		summary, err := s.dataProvider.GetFinancialSummary(ctx, userID)
		if err != nil {
			log.Printf("[ai] failed to fetch financial data for user %d: %v", userID, err)
		} else {
			financialData = s.formatFinancialData(summary)
		}
	}

	systemPrompt := fmt.Sprintf(`Anda adalah asisten keuangan pribadi untuk aplikasi Finance-GO. 

DATA PENGGUNA:
- Nama: %s
- ID: %d

%s

ATURAN:
1. Anda HANYA boleh menjawab pertanyaan seputar keuangan, analisis finansial, budgeting, utang, investasi, dan fitur-fitur di aplikasi Finance-GO.
2. JANGAN menjawab pertanyaan di luar topik keuangan (misal: "nama kamu siapa", "kamu bisa apa", dll). Beri tahu user bahwa Anda hanya fokus pada analisis keuangan.
3. Gunakan DATA KEUANGAN di atas untuk memberikan jawaban yang PERSONAL dan akurat.
4. Berikan jawaban yang singkat, jelas, dan actionable (bisa langsung dilakukan).
5. Gunakan bahasa Indonesia default, kecuali user bertanya dalam bahasa Inggris.
6. Rekomendasi harus spesifik dan praktis, berdasarkan data keuangan yang tersedia.
7. Jika data tertentu tidak tersedia, sampaikan dengan jujur.
8. Jangan menyebut Anda adalah AI atau bot - cukup bantu saja.
9. Maksimal 3-4 kalimat per jawaban.`, userName, userID, financialData)

	userPrompt := fmt.Sprintf(`Pertanyaan: %s

BANTUAN: Gunakan data keuangan pengguna yang sudah disediakan untuk memberikan analisis yang personal dan akurat.`, message)

	reqBody := deepseekRequest{
		Model: s.model,
		Messages: []deepseekMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   600,
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

	log.Printf("[ai] DeepSeek response status=%d", resp.StatusCode)

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

func (s *Service) formatFinancialData(summary FinancialSummary) string {
	var b strings.Builder
	b.WriteString("DATA KEUANGAN SAAT INI:\n")

	b.WriteString(fmt.Sprintf("- Total saldo: Rp%.0f\n", summary.TotalBalance))
	b.WriteString(fmt.Sprintf("- Pemasukan bulan ini: Rp%.0f\n", summary.MonthlyIncome))
	b.WriteString(fmt.Sprintf("- Pengeluaran bulan ini: Rp%.0f\n", summary.MonthlyExpense))
	b.WriteString(fmt.Sprintf("- Pengeluaran konsumsi: Rp%.0f\n", summary.ConsumptionExpense))
	b.WriteString(fmt.Sprintf("- Pembayaran utang: Rp%.0f\n", summary.DebtRepayment))
	b.WriteString(fmt.Sprintf("- Arus kas bersih: Rp%.0f\n", summary.NetCashflow))
	b.WriteString(fmt.Sprintf("- Rasio tabungan: %.1f%%\n", summary.SavingsRate))
	b.WriteString(fmt.Sprintf("- Rasio pengeluaran: %.1f%%\n", summary.ExpenseRatio))

	if summary.DebtRemaining > 0 {
		b.WriteString(fmt.Sprintf("- Total utang: Rp%.0f\n", summary.DebtTotal))
		b.WriteString(fmt.Sprintf("- Sisa utang: Rp%.0f\n", summary.DebtRemaining))
		b.WriteString(fmt.Sprintf("- Progress pelunasan: %.1f%%\n", summary.DebtCompletionRate))
	}

	if summary.BudgetUsage > 0 {
		b.WriteString(fmt.Sprintf("- Penggunaan budget: %.1f%%\n", summary.BudgetUsage))
		b.WriteString(fmt.Sprintf("- Sisa budget: Rp%.0f\n", summary.BudgetRemaining))
	}

	if len(summary.CategoryBreakdown) > 0 {
		b.WriteString("- Kategori pengeluaran terbesar:\n")
		for i, cat := range summary.CategoryBreakdown {
			if i >= 5 {
				break
			}
			b.WriteString(fmt.Sprintf("  • %s: Rp%.0f (%.1f%%)\n", cat.Category, cat.Amount, cat.Percentage))
		}
	}

	b.WriteString(fmt.Sprintf("- Total transaksi bulan ini: %d\n", summary.RecentTransactions))

	return b.String()
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
