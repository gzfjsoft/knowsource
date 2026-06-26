package constants

const (
	// MaxUploadSize 10MB in bytes
	MaxUploadSize = 10 << 20

	// MaxUploadSize 50MB in bytes
	KnowledgeMaxUploadSize = 10 << 50

	// FileTypeImage represents image file type
	FileTypeImage = "1"
	// FileTypeVideo represents video file type
	FileTypeVideo = "2"
	// FileTypeFile represents other file type
	FileTypeFile = "3"
	// FileTypeFile represents other file type
	FileTypeKnowledge = "4"
	// FileTypeRawDocuments represents raw documents file type
	FileTypeRawDocuments = "5"
)

// AllowedFileTypes maps file type to allowed extensions
var AllowedFileTypes = map[string][]string{
	FileTypeImage:        {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
	FileTypeVideo:        {".mp4", ".mov", ".avi", ".mkv", ".flv", ".wmv"},
	FileTypeFile:         {".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".zip", ".rar"},
	FileTypeKnowledge:    {".xls", ".xlsx", ".zip", ".csv"},
	FileTypeRawDocuments: {".docx", ".doc", ".pdf", ".txt", ".zip", ".md", ".xlsx"},
}
