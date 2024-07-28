package types

import (
	"time"
)

type LogFormat string
type LogLevel string
type LogOutput string

type DebuggingSettings struct {
	LogLevel  LogLevel  `json:"log_level"`
	LogFormat LogFormat `json:"log_format"`
	LogOutput LogOutput `json:"log_output"`
}

type GeneralSettings struct {
	ApplicationName string `json:"application_name"`
}

type SearchSettings struct {
	PerPage int `json:"per_page"`
}

type Settings struct {
	General   GeneralSettings   `json:"general"`
	Debugging DebuggingSettings `json:"debugging"`
	Search    SearchSettings    `json:"search"`
}

type AppSettings struct {
	ID            int       `db:"id"`
	Settings      Settings  `db:"settings"`
	LastUpdatedAt time.Time `db:"last_updated_at"`
}
