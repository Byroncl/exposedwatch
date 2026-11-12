package domain

// Stable lowercase names for firewall verdicts.
const (
	firewallNameAllowed         = "allowed"
	firewallNameRestricted      = "restricted"
	firewallNameBlocked         = "blocked"
	firewallNameUnknown         = "unknown"
	firewallNameNotExposed      = "not-exposed"
	firewallNameDockerPublished = "docker-published"
)

// Firewall verdicts. The zero value is FirewallUnknown so that a missing
// verdict can never be read as an assurance.
const (
	FirewallUnknown FirewallVerdict = iota
	FirewallAllowed
	FirewallRestricted
	FirewallBlocked
	FirewallNotExposed
	FirewallDockerPublished
)

// Address families of a rule set.
const (
	FamilyIPv4 AddressFamily = "ipv4"
	FamilyIPv6 AddressFamily = "ipv6"
)

// Well-known tables.
const (
	TableFilter TableName = "filter"
	TableNat    TableName = "nat"
	TableRaw    TableName = "raw"
	TableMangle TableName = "mangle"
)

// Well-known built-in chains.
const (
	ChainInput       ChainName = "INPUT"
	ChainForward     ChainName = "FORWARD"
	ChainOutput      ChainName = "OUTPUT"
	ChainPrerouting  ChainName = "PREROUTING"
	ChainPostrouting ChainName = "POSTROUTING"
)

// Well-known targets and chain policies.
const (
	TargetAccept   TargetName = "ACCEPT"
	TargetDrop     TargetName = "DROP"
	TargetReject   TargetName = "REJECT"
	TargetReturn   TargetName = "RETURN"
	TargetLog      TargetName = "LOG"
	TargetNflog    TargetName = "NFLOG"
	TargetUlog     TargetName = "ULOG"
	TargetDnat     TargetName = "DNAT"
	TargetRedirect TargetName = "REDIRECT"
	TargetTproxy   TargetName = "TPROXY"
	TargetNfqueue  TargetName = "NFQUEUE"
	TargetQueue    TargetName = "QUEUE"
)

// Connection tracking states.
const (
	StateNew         ConntrackState = "NEW"
	StateEstablished ConntrackState = "ESTABLISHED"
	StateRelated     ConntrackState = "RELATED"
	StateInvalid     ConntrackState = "INVALID"
	StateUntracked   ConntrackState = "UNTRACKED"
)

// Address types of the addrtype match.
const (
	AddressTypeLocal     AddressType = "LOCAL"
	AddressTypeMulticast AddressType = "MULTICAST"
	AddressTypeBroadcast AddressType = "BROADCAST"
	AddressTypeAnycast   AddressType = "ANYCAST"
)

// TCP flags of the tcp-flags match.
const (
	TCPFlagFin TCPFlag = "FIN"
	TCPFlagSyn TCPFlag = "SYN"
	TCPFlagRst TCPFlag = "RST"
	TCPFlagPsh TCPFlag = "PSH"
	TCPFlagAck TCPFlag = "ACK"
	TCPFlagUrg TCPFlag = "URG"
)

// Prefix lengths that make an address prefix lie wholly inside loopback space.
const (
	loopbackMinimumBitsIPv4 = 8
	loopbackMinimumBitsIPv6 = 128
)

// Backend reported when it cannot be determined.
const BackendUnknown = "unknown"

// Conntrack and security targets that can change how inbound traffic is
// tracked or classified before the filter table.
const (
	TargetCt          TargetName = "CT"
	TargetNotrack     TargetName = "NOTRACK"
	TargetSynproxy    TargetName = "SYNPROXY"
	TargetSecmark     TargetName = "SECMARK"
	TargetConnsecmark TargetName = "CONNSECMARK"
	// TargetUnparseable stands in for the target of a rule line that could
	// not be read, so it can never be mistaken for a harmless counting rule.
	TargetUnparseable TargetName = "UNPARSEABLE"
)

// What is known about native nftables rules next to the iptables rules.
const (
	NftablesNotChecked NftablesState = iota
	NftablesNone
	NftablesNative
	NftablesProbeFailed
)
