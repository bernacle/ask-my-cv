package message

type Citation struct {
	SourceTitle string `json:"source_title"`
	SourcePath  string `json:"source_path"`
	ChunkIndex  int32  `json:"chunk_index"`
}

type TokenUsage struct {
	InputTokens  int `json:"input"`
	OutputTokens int `json:"output"`
	TotalTokens  int `json:"total"`
}
