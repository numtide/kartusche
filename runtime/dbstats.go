package runtime

import "time"

type DBStats struct {
	FreePageN     int `json:"free_pages" yaml:"free_pages"`
	PendingPageN  int `json:"pending_pages" yaml:"pending_pages"`
	FreeAlloc     int `json:"free_pages_bytes" yaml:"free_pages_bytes"`
	FreelistInuse int `json:"free_list_bytes" yaml:"free_list_bytes"`

	TxN     int `json:"started_read_transactions" yaml:"started_read_transactions"`
	OpenTxN int `json:"open_read_transactions" yaml:"open_read_transactions"`

	TxStats TxStats // global, ongoing stats.
}

type TxStats struct {
	PageCount int `json:"page_allocations" yaml:"page_allocations"`
	PageAlloc int `json:"page_allocations_bytes" yaml:"page_allocations_bytes"`

	// Cursor statistics.
	CursorCount int `json:"cursor_count" yaml:"cursor_count"`

	// Node statistics
	NodeCount int `json:"node_count" yaml:"node_count"`
	NodeDeref int `json:"node_dereferences" yaml:"node_dereferences"`

	// Rebalance statistics.
	Rebalance     int           `json:"node_rebalances" yaml:"node_rebalances"`
	RebalanceTime time.Duration `json:"node_rebalancing_time" yaml:"node_rebalancing_time"`

	// Split/Spill statistics.
	Split     int           `json:"node_splits" yaml:"node_splits"`
	Spill     int           `json:"node_spills" yaml:"node_spills"`
	SpillTime time.Duration `json:"node_spilling_time" yaml:"node_spilling_time"`

	// Write statistics.
	Write     int           `json:"writes" yaml:"writes"`
	WriteTime time.Duration `json:"writing_time" yaml:"writing_time"`
}
