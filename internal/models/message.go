package models

import "time"

type Message struct {
	ID          int          `json:"id" db:"id"`                    
	Username    string       `json:"username" db:"username"`        
	Content     string       `json:"text" db:"content"`             
	SendAt      time.Time    `json:"created_at" db:"send_at"`       
	Attachments []Attachment `json:"attachments,omitempty"`         
}
