package models

type NGINX_vhosts struct {
	Listen          string
	ServerName      string `yaml:"server_name"`
	Root            string
	Index           string
	ErrorPage       string `yaml:"error_page"`
	AccessLog       string `yaml:"access_log"`
	ErrorLog        string `yaml:"error_log"`
	ExtraParameters string `yaml:"extra_parameters"`
	UUID            string
}

type Nginx_yaml struct {
	VHosts []NGINX_vhosts `yaml:"nginx_vhosts"`
}
