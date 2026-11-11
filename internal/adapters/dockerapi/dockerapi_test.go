package dockerapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Byroncl/exposedwatch/internal/domain"
	"github.com/Byroncl/exposedwatch/internal/ports"
)

var _ ports.PortSource = (*Client)(nil)

const (
	containersFixture   = "testdata/containers.json"
	temporaryDirPrefix  = "dk"
	temporarySocketName = "docker.sock"
	missingSocketName   = "missing.sock"
	invalidJSONBody     = "not json"
	contentTypeHeader   = "Content-Type"
	jsonContentType     = "application/json"
)

// serveOnUnixSocket starts handler on a temporary unix socket. The directory
// is created outside t.TempDir because unix socket paths are length-limited.
func serveOnUnixSocket(testRun *testing.T, handler http.Handler) SocketPath {
	testRun.Helper()
	socketDirectory, err := os.MkdirTemp("", temporaryDirPrefix)
	if err != nil {
		testRun.Fatal(err)
	}
	testRun.Cleanup(func() { _ = os.RemoveAll(socketDirectory) })
	socketPath := filepath.Join(socketDirectory, temporarySocketName)
	listener, err := net.Listen(unixNetwork, socketPath)
	if err != nil {
		testRun.Fatal(err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	testRun.Cleanup(server.Close)
	return SocketPath(socketPath)
}

func TestListPorts(testRun *testing.T) {
	fixtureContent, err := os.ReadFile(containersFixture)
	if err != nil {
		testRun.Fatal(err)
	}
	socketPath := serveOnUnixSocket(testRun, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/containers/json" || request.Method != http.MethodGet {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set(contentTypeHeader, jsonContentType)
		_, _ = writer.Write(fixtureContent)
	}))
	listedPorts, err := New(socketPath).ListPorts(context.Background())
	if err != nil {
		testRun.Fatal(err)
	}
	expectedPorts := []domain.Port{
		{Number: 8080, Protocol: domain.TCP, BindAddress: "0.0.0.0", Source: domain.SourceDocker, Container: "web"},
		{Number: 8080, Protocol: domain.TCP, BindAddress: "::", Source: domain.SourceDocker, Container: "web"},
		{Number: 5432, Protocol: domain.TCP, BindAddress: "127.0.0.1", Source: domain.SourceDocker, Container: "db"},
		{Number: 53, Protocol: domain.UDP, BindAddress: "0.0.0.0", Source: domain.SourceDocker, Container: "dns"},
	}
	if len(listedPorts) != len(expectedPorts) {
		testRun.Fatalf("got %d ports, want %d: %+v", len(listedPorts), len(expectedPorts), listedPorts)
	}
	for portIndex, expectedPort := range expectedPorts {
		if listedPorts[portIndex] != expectedPort {
			testRun.Errorf("port[%d] = %+v, want %+v", portIndex, listedPorts[portIndex], expectedPort)
		}
	}
}

func TestListPortsSocketMissing(testRun *testing.T) {
	missingSocket := SocketPath(filepath.Join(testRun.TempDir(), missingSocketName))
	_, err := New(missingSocket).ListPorts(context.Background())
	if !errors.Is(err, ErrDockerUnavailable) {
		testRun.Fatalf("err = %v, want ErrDockerUnavailable", err)
	}
}

func TestListPortsBadStatusAndJSON(testRun *testing.T) {
	responseStatus := http.StatusInternalServerError
	responseBody := ""
	socketPath := serveOnUnixSocket(testRun, http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(responseStatus)
		_, _ = writer.Write([]byte(responseBody))
	}))
	client := New(socketPath)
	if _, err := client.ListPorts(context.Background()); err == nil || errors.Is(err, ErrDockerUnavailable) {
		testRun.Fatalf("500: err = %v, want non-sentinel error", err)
	}
	responseStatus, responseBody = http.StatusOK, invalidJSONBody
	if _, err := client.ListPorts(context.Background()); err == nil {
		testRun.Fatal("expected JSON decode error")
	}
}

func TestListPortsContextCancelled(testRun *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := New(DefaultSocketPath).ListPorts(ctx)
	if !errors.Is(err, context.Canceled) {
		testRun.Fatalf("err = %v, want context.Canceled", err)
	}
}
