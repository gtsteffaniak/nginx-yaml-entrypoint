package main

type NginxConfig struct {
	LimitReqZones    []LimitReqZone      `yaml:"limit_req_zones"`
	ProxyCachePath   []ProxyCachePath    `yaml:"proxy_cache_path"`
	LimitReqLogLevel string              `yaml:"limit_req_log_level"`
	LimitReqStatus   int                 `yaml:"limit_req_status"`
	Servers          []NginxServer       `yaml:"servers"`
	TemplateConfigs  map[string]Location `yaml:"templateConfigs"`
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
	Path           string            `yaml:"path"`
	Configs        map[string]string `yaml:"configs"`
	AddHeader      map[string]string `yaml:"add_header"`
	ProxySetHeader map[string]string `yaml:"proxy_set_header"`
	LimitReqZone   LimitReqZone      `yaml:"limit_req"`
	ProxyPass      string            `yaml:"proxy_pass"`
	Conditions     []Condition       `yaml:"conditions"`
	ApplyTemplates []string          `yaml:"applyTemplates"`
}

type Condition struct {
	If   string   `yaml:"if"`
	Then []string `yaml:"then"`
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
	Locations            []Location        `yaml:"locations"`
	CustomConfig         []string          `yaml:"custom_config"`
	AddHeader            map[string]string `yaml:"add_header"`
	ProxySetHeader       map[string]string `yaml:"proxy_set_header"`
	Configs              map[string]string `yaml:"configs"`
	Defaults             Location          `yaml:"defaults"`
	ApplyTemplates       []string          `yaml:"applyTemplates"`
}
