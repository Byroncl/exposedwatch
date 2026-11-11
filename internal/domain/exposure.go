package domain

// Exposure describes how far a bind address reaches.
// Values are ordered from least to most exposed.
type Exposure int

// String returns a stable lowercase name.
func (exposure Exposure) String() string {
	switch exposure {
	case Loopback:
		return exposureNameLoopback
	case Private:
		return exposureNamePrivate
	case AllInterfaces:
		return exposureNameAllInterfaces
	case Public:
		return exposureNamePublic
	default:
		return exposureNameUnknown
	}
}
