package nmapscan

import (
	"encoding/xml"
	"fmt"
)

// openPort is an open TCP port found on one scanned address.
type openPort struct {
	Address string
	Number  int
}

type scanRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []scanHost `xml:"host"`
}

type scanHost struct {
	Addresses []hostAddress `xml:"address"`
	Ports     []hostPort    `xml:"ports>port"`
}

type hostAddress struct {
	Address string `xml:"addr,attr"`
	Type    string `xml:"addrtype,attr"`
}

type hostPort struct {
	Protocol string    `xml:"protocol,attr"`
	Number   int       `xml:"portid,attr"`
	State    portState `xml:"state"`
}

type portState struct {
	State string `xml:"state,attr"`
}

// parseResult extracts the open TCP ports of every host from `nmap -oX -`.
func parseResult(output []byte) ([]openPort, error) {
	var run scanRun
	if err := xml.Unmarshal(output, &run); err != nil {
		return nil, fmt.Errorf(invalidOutputFormat, ErrInvalidOutput, err)
	}
	if run.XMLName.Local != xmlRootElement {
		return nil, fmt.Errorf(invalidOutputFormat, ErrInvalidOutput, run.XMLName.Local)
	}
	var found []openPort
	for _, host := range run.Hosts {
		address := hostIPAddress(host)
		if address == "" {
			continue
		}
		for _, port := range host.Ports {
			if port.State.State == stateOpen && port.Protocol == protocolTCP {
				found = append(found, openPort{Address: address, Number: port.Number})
			}
		}
	}
	return found, nil
}

// hostIPAddress picks the IP address, ignoring MAC addresses.
func hostIPAddress(host scanHost) string {
	for _, address := range host.Addresses {
		if address.Type == addressTypeIPv4 || address.Type == addressTypeIPv6 {
			return address.Address
		}
	}
	return ""
}
