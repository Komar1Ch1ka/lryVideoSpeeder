package h5p

type Settings struct {
	TargetProgress  int     `json:"target_progress"`
	StepProgress    int     `json:"step_progress"`
	IntervalSeconds float64 `json:"interval_seconds"`
	MaxConcurrent   int     `json:"max_concurrent"`
}

type Config struct {
	Cookie   string   `json:"cookie"`
	Sesskey  string   `json:"sesskey"`
	Settings Settings `json:"settings"`
}

type Course struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type ProgressRequest struct {
	Index      int          `json:"index"`
	Methodname string       `json:"methodname"`
	Args       ProgressArgs `json:"args"`
}

type ProgressArgs struct {
	Time     int     `json:"time"`
	Finish   int     `json:"finish"`
	Cmid     string  `json:"cmid"`
	Total    int     `json:"total"`
	Progress float64 `json:"progress"`
}

type ProgressResponse struct {
	Error     bool                   `json:"error"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Exception *struct {
		Message string `json:"message"`
	} `json:"exception,omitempty"`
}
