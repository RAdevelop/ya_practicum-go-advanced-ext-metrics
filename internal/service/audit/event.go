package audit

// Event — событие аудита.
type Event struct {
	TS      int64    `json:"ts"`
	Metrics []string `json:"metrics"`
	IP      string   `json:"ip_address"`
}
