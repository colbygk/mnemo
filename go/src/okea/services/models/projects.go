package models

type Project struct {
	Title        string `json:"title"`
	UUID         string `json:"uuid"`
	OwnerUUID    string `json:"owner_uuid"`
	FQDN         string `json:"fqdn"`
	Enabled      bool   `json:"enabled"`
	CNameTarget  string `json:"cname_target"`
	UpstreamPort int    `json:"upstream_port"`
	UpstreamHost string `json:"upstream_host"`
}
