package models

type HealthStatus string

const (
	StatusUp   HealthStatus = "UP"
	StatusDown HealthStatus = "DOWN"
)

type Check struct {
	Name   string                 `json:"name"`
	Status HealthStatus           `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type HealthResponse struct {
	Status HealthStatus `json:"status"`
	Checks []Check      `json:"checks"`
}
