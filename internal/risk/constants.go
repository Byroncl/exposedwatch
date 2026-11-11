package risk

// Well-known service ports referenced by the rule table.
const (
	PortSSH           = 22
	PortTelnet        = 23
	PortMySQL         = 3306
	PortPostgreSQL    = 5432
	PortRedis         = 6379
	PortMongoDB       = 27017
	PortElasticsearch = 9200
	PortMemcached     = 11211
	PortDockerAPI     = 2375
	PortRDP           = 3389
	PortSMB           = 445
	PortNetBIOS       = 139
)

// Rule identifiers.
const (
	ruleIDGeneric   = "generic-exposed"
	ruleIDSSH       = "ssh-exposed"
	ruleIDTelnet    = "telnet-exposed"
	ruleIDDatabase  = "database-exposed"
	ruleIDDockerAPI = "docker-api-exposed"
	ruleIDRDP       = "rdp-exposed"
	ruleIDSMB       = "smb-exposed"
)

// Rule titles.
const (
	titleGeneric   = "Port listening on all interfaces"
	titleSSH       = "SSH exposed on all interfaces"
	titleTelnet    = "Telnet exposed on all interfaces"
	titleDatabase  = "Database or cache reachable beyond loopback"
	titleDockerAPI = "Unauthenticated Docker API exposed"
	titleRDP       = "RDP exposed on all interfaces"
	titleSMB       = "SMB exposed on all interfaces"
)

// Rule recommendations.
const (
	recommendationGeneric   = "Bind the service to 127.0.0.1 or a private address unless it must be reachable externally."
	recommendationSSH       = "Restrict SSH to trusted networks, disable password login and use key-based auth."
	recommendationTelnet    = "Disable telnet and use SSH instead; it transmits credentials in clear text."
	recommendationDatabase  = "Bind the datastore to 127.0.0.1 or an internal network and require authentication."
	recommendationDockerAPI = "Never expose the Docker daemon over plain TCP; use the unix socket or TLS with client certificates."
	recommendationRDP       = "Put RDP behind a VPN or gateway and enable network level authentication."
	recommendationSMB       = "Block SMB at the firewall and restrict it to the local network."
)
