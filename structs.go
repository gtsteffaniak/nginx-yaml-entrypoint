package main

type NginxConfig struct {
	LimitReqZones    []LimitReqZone   `yaml:"limit_req_zones"`
	ProxyCachePath   []ProxyCachePath `yaml:"proxy_cache_path"`
	LimitReqLogLevel string           `yaml:"limit_req_log_level"`
	LimitReqStatus   int              `yaml:"limit_req_status"`
	Servers          []NginxServer    `yaml:"servers"`
}
type LimitReqZone struct {
	Name  string `yaml:"name"`
	Key   string `yaml:"key"`
	Zone  string `yaml:"zone"`
	Rate  string `yaml:"rate"`
	Burst int    `yaml:"burst"`
}
type ProxyCachePath struct {
	Path     string `yaml:"path"`
	Levels   string `yaml:"levels"`
	KeysZone string `yaml:"keys_zone"`
	MaxSize  string `yaml:"max_size"`
}

type Location struct {
	Path         string            `yaml:"path"`
	Configs      map[string]string `yaml:"configs"`
	LimitReqZone LimitReqZone      `yaml:"limit_req"`
	ProxyPass    string            `yaml:"proxy_pass"`
}

type Servers struct {
	Listen    int        `yaml:"listen"`
	Locations []Location `yaml:"locations"`
}

type NginxServer struct {
	Listen               string            `yaml:"listen"`
	Name                 string            `yaml:"name"`
	ServerName           string            `yaml:"server_name"`
	ServerNameInRedirect string            `yaml:"server_name_in_redirect"`
	LimitReqZone         LimitReqZone      `yaml:"limit_req"`
	SslCertificate       string            `yaml:"ssl_certificate"`
	SslCertificateKey    string            `yaml:"ssl_certificate_key"`
	Locations            []Location        `yaml:"locations"`
	CustomConfig         []string          `yaml:"custom_config"`
	AddHeader            map[string]string `yaml:"add_header"`
	ProxySetHeader       map[string]string `yaml:"proxy_set_header"`
	Configs              map[string]string `yaml:"configs"`
}
