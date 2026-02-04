package models

type Attachment struct {
	ID        int    `json:"id" db:"id"`                    
	MessageID int    `json:"-" db:"message_id"`             
	FilePath  string `json:"filepath" db:"file_path"`       
	FileType  string `json:"filetype" db:"file_type"`       
	FileSize  int64  `json:"filesize" db:"file_size"`       
}