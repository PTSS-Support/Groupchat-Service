package repositories

// MessageQueryOptions defines the parameters for querying messages
type MessageQueryOptions struct {
	Limit     int    // Maximum number of messages to return
	Cursor    string // Token for pagination
	Direction string // Direction of pagination (next/previous)
	Search    string // Search term for filtering messages
}
