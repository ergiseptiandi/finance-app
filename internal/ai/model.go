package ai

type AnalysisRequest struct {
	Message string `json:"message"`
}

type AnalysisResponse struct {
	Reply string `json:"reply"`
}

type ContextData struct {
	UserID    int64
	Name      string
	Email     string
	Message   string
	CreatedAt string
}

type UsageInfo struct {
	ChatCount int `json:"chat_count"`
	MaxChats  int `json:"max_chats"`
}
