package access_log

import "github.com/eolinker/eosc"

type Config struct {
	Output []eosc.RequireId `json:"http-entry" skill:"github.com/eolinker/goku/http-entry.http-entry.IOutput"`
}
