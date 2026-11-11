package nmapscan

import (
	"errors"
	"net"
	"testing"
)

func fakeAddresses(addresses ...string) AddressLister {
	return func() ([]net.Addr, error) {
		listed := make([]net.Addr, 0, len(addresses))
		for _, text := range addresses {
			ipAddress, network, err := net.ParseCIDR(text)
			if err != nil {
				return nil, err
			}
			network.IP = ipAddress
			listed = append(listed, network)
		}
		return listed, nil
	}
}

func TestLocalTargetsKeepsOnlyRoutableIPv4(testRun *testing.T) {
	lister := fakeAddresses(
		"127.0.0.1/8", "::1/128", "169.254.10.1/16", "fe80::1/64", "2001:db8::5/64",
		"192.168.1.10/24", "172.17.0.1/16", "192.168.1.10/24", "0.0.0.0/0", "224.0.0.1/4",
	)
	targets, err := LocalTargets(lister)
	if err != nil {
		testRun.Fatal(err)
	}
	if len(targets) != 2 || targets[0] != "192.168.1.10" || targets[1] != "172.17.0.1" {
		testRun.Fatalf("targets=%v", targets)
	}
}

func TestLocalTargetsWithoutUsableAddressIsError(testRun *testing.T) {
	_, err := LocalTargets(fakeAddresses("127.0.0.1/8", "fe80::1/64"))
	if !errors.Is(err, ErrNoTargets) {
		testRun.Fatalf("err=%v", err)
	}
}

func TestLocalTargetsPropagatesListerFailure(testRun *testing.T) {
	failure := errors.New("boom")
	_, err := LocalTargets(func() ([]net.Addr, error) { return nil, failure })
	if !errors.Is(err, failure) {
		testRun.Fatalf("err=%v", err)
	}
}
