package stats

type PluginStats struct {
	Count           int  `json:"count" yaml:"count"`
	Enterprise      bool `json:"enterprise,omitempty" yaml:"enterprise,omitempty"`
	OSS             bool `json:"oss,omitempty" yaml:"oss,omitempty"`
	CustomOrPartner bool `json:"customOrPartner,omitempty" yaml:"customOrPartner,omitempty"`
}

type ContentStatistics struct {
	WorkspaceOrCPName   string                 `json:"workspaceOrCPName" yaml:"workspaceOrCPName"`
	Services            int                    `json:"services" yaml:"services"`
	Routes              int                    `json:"routes" yaml:"routes"`
	Consumers           int                    `json:"consumers" yaml:"consumers"`
	ConsumerGroups      int                    `json:"consumerGroups" yaml:"consumerGroups"`
	Plugins             int                    `json:"plugins" yaml:"plugins"`
	Upstreams           int                    `json:"upstreams" yaml:"upstreams"`
	Certificates        int                    `json:"certificates" yaml:"certificates"`
	CACertificates      int                    `json:"caCertificates" yaml:"caCertificates"`
	Vaults              int                    `json:"vaults" yaml:"vaults"`
	TotalEntities       int                    `json:"totalEntities" yaml:"totalEntities"`
	ServicesPct         float64                `json:"servicesPct" yaml:"servicesPct"`
	RoutesPct           float64                `json:"routesPct" yaml:"routesPct"`
	ConsumersPct        float64                `json:"consumersPct" yaml:"consumersPct"`
	ConsumerGroupsPct   float64                `json:"consumerGroupsPct" yaml:"consumerGroupsPct"`
	PluginsPct          float64                `json:"pluginsPct" yaml:"pluginsPct"`
	UpstreamsPct        float64                `json:"upstreamsPct" yaml:"upstreamsPct"`
	CertificatesPct     float64                `json:"certificatesPct" yaml:"certificatesPct"`
	CACertificatesPct   float64                `json:"caCertificatesPct" yaml:"caCertificatesPct"`
	VaultsPct           float64                `json:"vaultsPct" yaml:"vaultsPct"`
	PluginsCountPerName map[string]PluginStats `json:"pluginsCountPerName" yaml:"pluginsCountPerName"`
	OSSPlugins          int                    `json:"ossPlugins" yaml:"ossPlugins"`
	EnterprisePlugins   int                    `json:"enterprisePlugins" yaml:"enterprisePlugins"`
	CustomPlugins       int                    `json:"customOrPartnerPlugins" yaml:"customOrPartnerPlugins"`
}
