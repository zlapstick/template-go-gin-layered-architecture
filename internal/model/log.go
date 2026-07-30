package model

type LogFile struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
}

type LogEntry struct {
	Level   string `json:"level,omitempty"`
	Module  string `json:"module,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Time    string `json:"time,omitempty"`
	UserID  string `json:"user_id,omitempty"`
	Error   string `json:"error,omitempty"`
	Path    string `json:"path,omitempty"`
}
