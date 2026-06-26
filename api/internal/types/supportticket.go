package types

// SupportTicket is used by ticket-related logic helpers.
// Kept in a separate file to avoid editing generated `types.go`.
type SupportTicket struct {
	TicketId    uint64   `json:"ticketId"`
	UserId      uint64   `json:"userId"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Images      string   `json:"images"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
	IsRead      uint64   `json:"isRead"`
}

