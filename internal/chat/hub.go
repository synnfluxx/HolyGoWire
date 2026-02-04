package chat

import (
	"TextMeByte/internal/models"
	"TextMeByte/internal/storage"
	"encoding/json"
	"sync"
)

type Hub struct {
	Clients   map[*Client]bool     
	Broadcast chan models.Message  
	Store     *storage.Storage     
	Mu        sync.RWMutex         
}

func NewHub(store *storage.Storage) *Hub {
	return &Hub{
		Clients:   make(map[*Client]bool),
		Broadcast: make(chan models.Message),
		Store:     store,
	}
}

func (h *Hub) Run() {
	for msg := range h.Broadcast {
		
		payload, _ := json.Marshal(msg)

		
		h.Mu.Lock()
		for c := range h.Clients {
			select {
			case c.Send <- payload:
				
			default:
				
				close(c.Send)
				delete(h.Clients, c)
			}
		}
		h.Mu.Unlock()
	}
}