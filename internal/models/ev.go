package models

// EVSpec represents the battery specifications for an electric vehicle
type EVSpec struct {
	Make       string  `json:"make"`
	Model      string  `json:"model"`
	Year       int     `json:"year"`
	Capacity   float64 `json:"capacity_kwh"` // Battery capacity in kWh
	Power      float64 `json:"power_kw"`     // Power output in kW
	Chemistry  string  `json:"chemistry"`    // Battery chemistry type
	Confidence float64 `json:"confidence"`   // Confidence score from similarity search
	Source     string  `json:"source"`       // Source of the data (e.g., "database", "llm")
}
