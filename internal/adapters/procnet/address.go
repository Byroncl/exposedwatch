package procnet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
)

// decodeAddress decodes the kernel hex form: each 32-bit word is printed in
// host (little-endian) byte order, so every 4-byte group is reversed.
func decodeAddress(addressHex string) (string, error) {
	addressBytes, err := hex.DecodeString(addressHex)
	if err != nil {
		return "", err
	}
	if len(addressBytes) != ipv4ByteLength && len(addressBytes) != ipv6ByteLength {
		return "", fmt.Errorf(unexpectedLengthMessage, len(addressBytes))
	}
	for wordStart := 0; wordStart < len(addressBytes); wordStart += addressWordByteSize {
		word := addressBytes[wordStart : wordStart+addressWordByteSize]
		word[0], word[1], word[2], word[3] = word[3], word[2], word[1], word[0]
	}
	address, valid := netip.AddrFromSlice(addressBytes)
	if !valid {
		return "", errors.New(invalidAddressMessage)
	}
	return address.String(), nil
}
