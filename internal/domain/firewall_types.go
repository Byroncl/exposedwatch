package domain

// AddressFamily distinguishes IPv4 from IPv6 rule sets.
type AddressFamily string

// TableName names an iptables table.
type TableName string

// ChainName names an iptables chain.
type ChainName string

// TargetName names a rule target (ACCEPT, DROP, a chain, ...) or a chain policy.
type TargetName string

// InterfaceName is a network interface name or pattern such as docker+.
type InterfaceName string

// ConntrackState is a connection tracking state.
type ConntrackState string

// AddressType is a kernel address type used by the addrtype match.
type AddressType string

// TCPFlag is one TCP header flag.
type TCPFlag string

// NftablesState says whether native nftables rules exist besides the iptables
// rules. The zero value means nobody checked.
type NftablesState int
