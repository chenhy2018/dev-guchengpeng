package api

type FileContent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Content string `json:"content"`
}
