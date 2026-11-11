package risk

import "github.com/Byroncl/exposedwatch/internal/domain"

// Rule flags a well-known service port once it is exposed at least as widely
// as MinExposure.
type Rule struct {
	ID             string
	Ports          []int
	MinExposure    domain.Exposure
	Severity       domain.Severity
	Title          string
	Recommendation string
}

// genericRule applies to any port exposed on all interfaces that no
// specific rule matched.
var genericRule = Rule{
	ID:             ruleIDGeneric,
	MinExposure:    domain.AllInterfaces,
	Severity:       domain.Info,
	Title:          titleGeneric,
	Recommendation: recommendationGeneric,
}

// builtinRules is the extensible rule table. Add a row to add a rule.
var builtinRules = []Rule{
	{ruleIDSSH, []int{PortSSH}, domain.AllInterfaces, domain.Medium, titleSSH, recommendationSSH},
	{ruleIDTelnet, []int{PortTelnet}, domain.AllInterfaces, domain.Critical, titleTelnet, recommendationTelnet},
	{ruleIDDatabase, []int{PortMySQL, PortPostgreSQL, PortRedis, PortMongoDB, PortElasticsearch, PortMemcached},
		domain.Private, domain.High, titleDatabase, recommendationDatabase},
	{ruleIDDockerAPI, []int{PortDockerAPI}, domain.AllInterfaces, domain.Critical, titleDockerAPI, recommendationDockerAPI},
	{ruleIDRDP, []int{PortRDP}, domain.AllInterfaces, domain.High, titleRDP, recommendationRDP},
	{ruleIDSMB, []int{PortSMB, PortNetBIOS}, domain.AllInterfaces, domain.High, titleSMB, recommendationSMB},
}
