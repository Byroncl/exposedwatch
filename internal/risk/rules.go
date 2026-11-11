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
	ID:             "generic-exposed",
	MinExposure:    domain.AllInterfaces,
	Severity:       domain.Info,
	Title:          "Port listening on all interfaces",
	Recommendation: "Bind the service to 127.0.0.1 or a private address unless it must be reachable externally.",
}

// builtinRules is the extensible rule table. Add a row to add a rule.
var builtinRules = []Rule{
	{"ssh-exposed", []int{22}, domain.AllInterfaces, domain.Medium,
		"SSH exposed on all interfaces",
		"Restrict SSH to trusted networks, disable password login and use key-based auth."},
	{"telnet-exposed", []int{23}, domain.AllInterfaces, domain.Critical,
		"Telnet exposed on all interfaces",
		"Disable telnet and use SSH instead; it transmits credentials in clear text."},
	{"database-exposed", []int{3306, 5432, 6379, 27017, 9200, 11211}, domain.Private, domain.High,
		"Database or cache reachable beyond loopback",
		"Bind the datastore to 127.0.0.1 or an internal network and require authentication."},
	{"docker-api-exposed", []int{2375}, domain.AllInterfaces, domain.Critical,
		"Unauthenticated Docker API exposed",
		"Never expose the Docker daemon over plain TCP; use the unix socket or TLS with client certificates."},
	{"rdp-exposed", []int{3389}, domain.AllInterfaces, domain.High,
		"RDP exposed on all interfaces",
		"Put RDP behind a VPN or gateway and enable network level authentication."},
	{"smb-exposed", []int{445, 139}, domain.AllInterfaces, domain.High,
		"SMB exposed on all interfaces",
		"Block SMB at the firewall and restrict it to the local network."},
}
