package media

type UploadResult struct {
	Path     string `json:"path"`
	URL      string `json:"url"`
	Dir      string `json:"dir"`
	FileName string `json:"file_name"`
}
