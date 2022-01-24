package fetcher

import "context"

// Response holds the config response received from the config provider
type Response struct {
	FileName string `json:"filename"`
	FileData []byte `json:"fileData"`
}

// Fetcher interface must be implemented by the config provider
type Fetcher interface {
	Fetch(ctx context.Context) (*Response, error)
}
