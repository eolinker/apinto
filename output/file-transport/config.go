package file_transport

//Config filelog-Transporter所需配置
type Config struct {
	Dir    string
	File   string
	Expire int
	Period LogPeriod
}

func (c *Config) IsUpdate(cfg *Config) bool {
	if cfg.File != c.File || cfg.Dir != c.Dir || cfg.Period != c.Period || cfg.Expire != c.Expire {
		return true
	}
	return false
}
