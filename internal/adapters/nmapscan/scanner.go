// Package nmapscan scans this machine from the network side with nmap and
// reports the open TCP ports it finds. nmap is executed without a shell,
// with fixed flags and strictly validated targets.
package nmapscan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Options configures a Scanner. Zero values select safe defaults.
type Options struct {
	// NmapPath is the executable; empty means BinaryName looked up in PATH.
	NmapPath string
	// Targets to scan; empty means this host's own addresses.
	Targets TargetList
	// Ports to scan; empty means nmap's top 1000 ports.
	Ports PortSpec
	// Timeout bounds the scan; zero means DefaultTimeout.
	Timeout time.Duration
	// Runner executes nmap; nil means the real os/exec runner.
	Runner CommandRunner
	// AddressLister supplies local addresses; nil means net.InterfaceAddrs.
	AddressLister AddressLister
}

// Scanner implements ports.PortSource over nmap.
type Scanner struct {
	options Options
	// processLimit overrides the hard process deadline; zero derives it from
	// the timeout. It exists so tests can exercise the kill path quickly.
	processLimit time.Duration
}

// New builds a Scanner, filling in defaults.
func New(options Options) *Scanner {
	if options.NmapPath == "" {
		options.NmapPath = BinaryName
	}
	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
	}
	if options.Runner == nil {
		options.Runner = NewExecRunner()
	}
	if options.AddressLister == nil {
		options.AddressLister = net.InterfaceAddrs
	}
	return &Scanner{options: options}
}

// ValidateTimeout checks a scan timeout against the allowed bounds.
func ValidateTimeout(timeout time.Duration) error {
	if timeout < MinTimeout || timeout > MaxTimeout {
		return fmt.Errorf(invalidTimeoutFormat, ErrInvalidTimeout, timeout, MinTimeout, MaxTimeout)
	}
	return nil
}

// ListPorts scans the targets and returns their open TCP ports as external
// ports. All inputs are validated again before any process is started.
func (scanner *Scanner) ListPorts(ctx context.Context) ([]domain.Port, error) {
	targets, err := scanner.resolveTargets()
	if err != nil {
		return nil, err
	}
	if _, err := ParsePortSpec(string(scanner.options.Ports)); err != nil {
		return nil, err
	}
	if err := ValidateTimeout(scanner.options.Timeout); err != nil {
		return nil, err
	}

	limit := scanner.options.Timeout + processGrace
	if scanner.processLimit > 0 {
		limit = scanner.processLimit
	}
	runCtx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()

	result, err := scanner.options.Runner.Run(runCtx, scanner.options.NmapPath, scanner.buildArguments(targets))
	if err != nil {
		if ctx.Err() == nil && errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf(timeoutFormat, ErrScanTimedOut, scanner.options.Timeout)
		}
		return nil, err
	}
	if result.ExitCode != noExitCode {
		return nil, fmt.Errorf(nmapFailedFormat, ErrNmapFailed, result.ExitCode, boundedText(result.Stderr))
	}
	found, err := parseResult(result.Stdout)
	if err != nil {
		return nil, err
	}
	return toDomainPorts(found), nil
}

func (scanner *Scanner) resolveTargets() (TargetList, error) {
	if len(scanner.options.Targets) == 0 {
		return LocalTargets(scanner.options.AddressLister)
	}
	if err := scanner.options.Targets.Validate(); err != nil {
		return nil, err
	}
	return scanner.options.Targets, nil
}

// buildArguments returns the fixed nmap argument list. Targets are separate
// elements placed after the "--" terminator.
func (scanner *Scanner) buildArguments(targets TargetList) []string {
	arguments := []string{
		outputFlag, outputToStdout, skipHostDiscovery, skipDNSResolution, onlyOpenPorts,
		tcpConnectScan, timingFlag,
		hostTimeoutFlag, fmt.Sprintf(secondsFormat, int(scanner.options.Timeout/time.Second)),
	}
	if scanner.options.Ports == "" {
		arguments = append(arguments, topPortsFlag, defaultTopPorts)
	} else {
		arguments = append(arguments, portsFlag, string(scanner.options.Ports))
	}
	arguments = append(arguments, terminatorArgument)
	for _, target := range targets {
		arguments = append(arguments, string(target))
	}
	return arguments
}

func toDomainPorts(found []openPort) []domain.Port {
	converted := make([]domain.Port, 0, len(found))
	for _, port := range found {
		converted = append(converted, domain.Port{
			Number:      port.Number,
			Protocol:    domain.TCP,
			BindAddress: port.Address,
			Source:      domain.SourceExternal,
		})
	}
	return converted
}

// boundedText trims nmap's stderr and caps its length for error messages.
func boundedText(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if len(text) > maxStderrCharacters {
		text = text[:maxStderrCharacters]
	}
	return strconv.Quote(text)
}
