package main

import "fmt"

// NginxConfig represents the complete nginx configuration
type NginxConfig struct {
	// Global directives
	WorkerProcesses     string              `yaml:"worker_processes,omitempty"`
	WorkerConnections   int                 `yaml:"worker_connections,omitempty"`
	KeepaliveTimeout    string              `yaml:"keepalive_timeout,omitempty"`
	ClientMaxBodySize   string              `yaml:"client_max_body_size,omitempty"`
	ServerTokens        string              `yaml:"server_tokens,omitempty"`
	
	// Rate limiting
	LimitReqZones       []LimitReqZone      `yaml:"limit_req_zones,omitempty"`
	LimitReqLogLevel    string              `yaml:"limit_req_log_level,omitempty"`
	LimitReqStatus      int                 `yaml:"limit_req_status,omitempty"`
	
	// Caching
	ProxyCachePath      []ProxyCachePath    `yaml:"proxy_cache_path,omitempty"`
	
	// Compression
	Gzip                *GzipConfig         `yaml:"gzip,omitempty"`
	
	// Logging
	AccessLog           *LogConfig          `yaml:"access_log,omitempty"`
	ErrorLog            *LogConfig          `yaml:"error_log,omitempty"`
	
	// Load balancing
	Upstreams           []Upstream          `yaml:"upstreams,omitempty"`
	
	// Security
	SecurityHeaders     *SecurityHeaders    `yaml:"security_headers,omitempty"`
	
	// Servers and templates
	Servers             []NginxServer       `yaml:"servers"`
	TemplateConfigs     map[string]Location `yaml:"templateConfigs,omitempty"`
	
	// Custom global config for anything not covered above
	CustomGlobalConfig  []string            `yaml:"custom_global_config,omitempty"`
}

// LimitReqZone defines rate limiting zones
type LimitReqZone struct {
	Name  string `yaml:"name"`
	Key   string `yaml:"key"`
	Zone  string `yaml:"zone"`
	Rate  string `yaml:"rate"`
	Burst int    `yaml:"burst,omitempty"`
}

// ProxyCachePath defines proxy cache storage
type ProxyCachePath struct {
	Name         string `yaml:"name"`
	Path         string `yaml:"path"`
	Levels       string `yaml:"levels,omitempty"`
	KeysZone     string `yaml:"keys_zone"`
	MaxSize      string `yaml:"max_size,omitempty"`
	Inactive     string `yaml:"inactive,omitempty"`
	UseStaleError bool  `yaml:"use_stale_error,omitempty"`
}

// GzipConfig defines compression settings
type GzipConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Vary         bool     `yaml:"vary,omitempty"`
	Proxied      string   `yaml:"proxied,omitempty"`
	CompLevel    int      `yaml:"comp_level,omitempty"`
	MinLength    int      `yaml:"min_length,omitempty"`
	Types        []string `yaml:"types,omitempty"`
	Disable      string   `yaml:"disable,omitempty"`
}

// LogConfig defines logging configuration
type LogConfig struct {
	Path   string `yaml:"path,omitempty"`
	Format string `yaml:"format,omitempty"`
	Level  string `yaml:"level,omitempty"`
}

// Upstream defines load balancing upstream servers
type Upstream struct {
	Name        string           `yaml:"name"`
	Servers     []UpstreamServer `yaml:"servers"`
	Method      string           `yaml:"method,omitempty"`       // round_robin, ip_hash, least_conn, etc.
	KeepaliveConnections int     `yaml:"keepalive,omitempty"`
	KeepaliveRequests    int     `yaml:"keepalive_requests,omitempty"`
	KeepaliveTimeout     string  `yaml:"keepalive_timeout,omitempty"`
}

// UpstreamServer defines individual upstream server
type UpstreamServer struct {
	Address    string `yaml:"address"`
	Weight     int    `yaml:"weight,omitempty"`
	MaxFails   int    `yaml:"max_fails,omitempty"`
	FailTimeout string `yaml:"fail_timeout,omitempty"`
	Backup     bool   `yaml:"backup,omitempty"`
	Down       bool   `yaml:"down,omitempty"`
}

// SecurityHeaders defines common security headers
type SecurityHeaders struct {
	HSTS                   *HSTSConfig `yaml:"hsts,omitempty"`
	ContentSecurityPolicy  string      `yaml:"csp,omitempty"`
	XFrameOptions          string      `yaml:"x_frame_options,omitempty"`
	XContentTypeOptions    string      `yaml:"x_content_type_options,omitempty"`
	ReferrerPolicy         string      `yaml:"referrer_policy,omitempty"`
	PermissionsPolicy      string      `yaml:"permissions_policy,omitempty"`
	
	// CORS settings
	CORS                   *CORSConfig `yaml:"cors,omitempty"`
}

// HSTSConfig defines HSTS settings
type HSTSConfig struct {
	MaxAge            int  `yaml:"max_age"`
	IncludeSubdomains bool `yaml:"include_subdomains,omitempty"`
	Preload           bool `yaml:"preload,omitempty"`
}

// CORSConfig defines CORS settings
type CORSConfig struct {
	AllowOrigin      string   `yaml:"allow_origin,omitempty"`
	AllowMethods     []string `yaml:"allow_methods,omitempty"`
	AllowHeaders     []string `yaml:"allow_headers,omitempty"`
	ExposeHeaders    []string `yaml:"expose_headers,omitempty"`
	AllowCredentials bool     `yaml:"allow_credentials,omitempty"`
	MaxAge           int      `yaml:"max_age,omitempty"`
}

// Location defines a location block within a server
type Location struct {
	Path                string            `yaml:"path"`
	
	// Basic proxy settings
	ProxyPass           string            `yaml:"proxy_pass,omitempty"`
	ProxySetHeader      map[string]string `yaml:"proxy_set_header,omitempty"`
	ProxyPassHeader     []string          `yaml:"proxy_pass_header,omitempty"`
	ProxyHideHeader     []string          `yaml:"proxy_hide_header,omitempty"`
	ProxyIgnoreHeaders  []string          `yaml:"proxy_ignore_headers,omitempty"`
	
	// Proxy buffering
	ProxyBuffering      *bool             `yaml:"proxy_buffering,omitempty"`
	ProxyBuffers        string            `yaml:"proxy_buffers,omitempty"`
	ProxyBufferSize     string            `yaml:"proxy_buffer_size,omitempty"`
	ProxyBusyBuffersSize string           `yaml:"proxy_busy_buffers_size,omitempty"`
	
	// Proxy timeouts
	ProxyConnectTimeout string            `yaml:"proxy_connect_timeout,omitempty"`
	ProxySendTimeout    string            `yaml:"proxy_send_timeout,omitempty"`
	ProxyReadTimeout    string            `yaml:"proxy_read_timeout,omitempty"`
	
	// Caching
	ProxyCache          string            `yaml:"proxy_cache,omitempty"`
	ProxyCacheValid     map[string]string `yaml:"proxy_cache_valid,omitempty"`
	ProxyCacheKey       string            `yaml:"proxy_cache_key,omitempty"`
	ProxyCacheBypass    []string          `yaml:"proxy_cache_bypass,omitempty"`
	ProxyCacheLock      bool              `yaml:"proxy_cache_lock,omitempty"`
	
	// Headers
	AddHeader           map[string]string `yaml:"add_header,omitempty"`
	MoreSetHeaders      map[string]string `yaml:"more_set_headers,omitempty"`
	
	// Rate limiting
	LimitReqZone        LimitReqZone      `yaml:"limit_req,omitempty"`
	
	// Static files
	Root                string            `yaml:"root,omitempty"`
	Index               []string          `yaml:"index,omitempty"`
	TryFiles            string            `yaml:"try_files,omitempty"`
	Expires             string            `yaml:"expires,omitempty"`
	
	// Access control
	Allow               []string          `yaml:"allow,omitempty"`
	Deny                []string          `yaml:"deny,omitempty"`
	
	// Authentication
	AuthRequest         string            `yaml:"auth_request,omitempty"`
	AuthRequestSet      map[string]string `yaml:"auth_request_set,omitempty"`
	
	// WebDAV
	DavMethods          []string          `yaml:"dav_methods,omitempty"`
	DavAccess           string            `yaml:"dav_access,omitempty"`
	
	// Custom configurations
	Configs             map[string]string `yaml:"configs,omitempty"`
	Conditions          []Condition       `yaml:"conditions,omitempty"`
	ApplyTemplates      []string          `yaml:"applyTemplates,omitempty"`
	CustomConfig        []string          `yaml:"custom_config,omitempty"`
}

// Condition represents nginx if-statements
type Condition struct {
	If   string   `yaml:"if"`
	Then []string `yaml:"then"`
}

// NginxServer defines a server block
type NginxServer struct {
	// Basic server settings
	Listen      string `yaml:"listen"`
	ServerName  string `yaml:"server_name"`
	Name        string `yaml:"name,omitempty"`
	
	// SSL/TLS settings
	SSL         *SSLConfig `yaml:"ssl,omitempty"`
	
	// HTTP settings
	HTTP2       bool   `yaml:"http2,omitempty"`
	HTTP3       bool   `yaml:"http3,omitempty"`
	
	// Error handling
	ErrorPage   map[string]string `yaml:"error_page,omitempty"`
	
	// Redirects
	Return      string `yaml:"return,omitempty"`
	Rewrite     []RewriteRule `yaml:"rewrite,omitempty"`
	
	// Rate limiting
	LimitReqZone LimitReqZone `yaml:"limit_req,omitempty"`
	
	// Headers
	AddHeader      map[string]string `yaml:"add_header,omitempty"`
	ProxySetHeader map[string]string `yaml:"proxy_set_header,omitempty"`
	MoreSetHeaders map[string]string `yaml:"more_set_headers,omitempty"`
	
	// Locations
	Locations   []Location `yaml:"locations,omitempty"`
	Defaults    Location   `yaml:"defaults,omitempty"`
	
	// Templates
	ApplyTemplates []string `yaml:"applyTemplates,omitempty"`
	
	// Custom configurations  
	Configs        map[string]string `yaml:"configs,omitempty"`
	CustomConfig   []string          `yaml:"custom_config,omitempty"`
}

// SSLConfig defines SSL/TLS settings
type SSLConfig struct {
	Certificate    string   `yaml:"certificate"`
	CertificateKey string   `yaml:"certificate_key"`
	Protocols      []string `yaml:"protocols,omitempty"`
	Ciphers        string   `yaml:"ciphers,omitempty"`
	DHParam        string   `yaml:"dhparam,omitempty"`
	
	// Session settings
	SessionCache   string `yaml:"session_cache,omitempty"`
	SessionTimeout string `yaml:"session_timeout,omitempty"`
	SessionTickets bool   `yaml:"session_tickets,omitempty"`
	
	// OCSP
	StaplingEnabled bool   `yaml:"stapling,omitempty"`
	StaplingVerify  bool   `yaml:"stapling_verify,omitempty"`
	TrustedCert     string `yaml:"trusted_certificate,omitempty"`
	
	// Let's Encrypt
	ACMEChallengeRoot string `yaml:"acme_challenge_root,omitempty"`
}

// RewriteRule defines URL rewrite rules
type RewriteRule struct {
	Regex       string `yaml:"regex"`
	Replacement string `yaml:"replacement"`
	Flag        string `yaml:"flag,omitempty"` // last, break, redirect, permanent
}

// ValidationError represents configuration validation errors
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s' (value: '%s'): %s", ve.Field, ve.Value, ve.Message)
}
