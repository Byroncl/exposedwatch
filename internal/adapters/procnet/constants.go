package procnet

// SocketState is the hexadecimal connection state column of /proc/net/tcp.
type SocketState string

// ProcFileName is a file or directory name inside a /proc tree.
type ProcFileName string

// SocketInode is the kernel inode identifying one socket.
type SocketInode string

// DefaultProcRoot is where the proc filesystem is normally mounted.
const DefaultProcRoot = "/proc"

// SocketStateListen is the TCP LISTEN state code.
const SocketStateListen SocketState = "0A"

// Names inside the proc tree.
const (
	netDirectoryName ProcFileName = "net"
	fdDirectoryName  ProcFileName = "fd"
	commFileName     ProcFileName = "comm"

	tcpTableName  ProcFileName = "tcp"
	tcp6TableName ProcFileName = "tcp6"
	udpTableName  ProcFileName = "udp"
	udp6TableName ProcFileName = "udp6"

	processDirectoryGlob = "[0-9]*"
)

// Socket link format used by /proc/<pid>/fd entries: "socket:[<inode>]".
const (
	socketLinkPrefix = "socket:["
	socketLinkSuffix = "]"
)

// Column layout of a socket row.
const (
	minimumColumnCount      = 10
	localAddressColumnIndex = 1
	stateColumnIndex        = 3
	inodeColumnIndex        = 9
	addressPortSeparator    = ":"
)

// Numeric decoding parameters.
const (
	hexadecimalBase     = 16
	portBitSize         = 16
	unboundPortNumber   = 0
	ipv4ByteLength      = 4
	ipv6ByteLength      = 16
	addressWordByteSize = 4
)

// Error message formats.
const (
	noTableReadableMessage  = "procnet: no socket table readable: %w"
	unexpectedLengthMessage = "procnet: unexpected address length %d"
	invalidAddressMessage   = "procnet: invalid address"
)
