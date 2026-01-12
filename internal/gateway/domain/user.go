package domain

type UserInfo struct {
	Balance         uint32          `json:"balance"`
	Inventory       []InventoryItem `json:"inventory"`
	TransferHistory TransferHistory `json:"coinHistory"`
}

type TransferHistory struct {
	Received []ReceivedTransfer `json:"received"`
	Sent     []SentTransfer     `json:"sent"`
}

type InventoryItem struct {
	Name     string `json:"type"`
	Quantity uint32 `json:"quantity"`
}

type ReceivedTransfer struct {
	From   string `json:"fromUser"`
	Amount uint32 `json:"amount"`
}

type SentTransfer struct {
	To     string `json:"toUser"`
	Amount uint32 `json:"amount"`
}
