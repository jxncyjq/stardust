package components

import (
	"testing"

	"github.com/jxncyjq/stardust/app"
)

func TestHTTPServerComponentContract(t *testing.T) {
	component := HTTPServerComponent(nil)

	if component.Name() != "http_server" {
		t.Fatalf("Name() = %q, want http_server", component.Name())
	}
	if deps := component.Dependencies(); len(deps) != 1 || deps[0] != "logs" {
		t.Fatalf("Dependencies() = %#v, want [logs]", deps)
	}
}

func TestHTTPServerFromAppComponentContract(t *testing.T) {
	component := HTTPServerFromApp(app.New(nil), nil)

	if component.Name() != "http_server" {
		t.Fatalf("Name() = %q, want http_server", component.Name())
	}
	if deps := component.Dependencies(); len(deps) != 1 || deps[0] != "logs" {
		t.Fatalf("Dependencies() = %#v, want [logs]", deps)
	}
}

func TestGRPCServerComponentContract(t *testing.T) {
	component := GRPCServerComponent(nil)

	if component.Name() != "grpc_server" {
		t.Fatalf("Name() = %q, want grpc_server", component.Name())
	}
	deps := component.Dependencies()
	if len(deps) != 2 || deps[0] != "logs" || deps[1] != "tracing" {
		t.Fatalf("Dependencies() = %#v, want [logs tracing]", deps)
	}
}
