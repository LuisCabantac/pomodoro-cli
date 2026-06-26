package stats

type StatList struct {
	Stats []Stat `json:"stats"`
}

type Stat struct {
	Date        string `json:"date"`
	PresetID    string `json:"presetId"`
	DurationMin int    `json:"duration_min"`
	Count       int    `json:"count"`
}
