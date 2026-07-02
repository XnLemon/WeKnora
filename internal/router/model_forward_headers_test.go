package router

import (
	"os"
	"strings"
	"testing"
)

func TestModelHeaderForwardingRegisteredBeforePublicModelRoutes(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	routerSource := string(source)

	forwardingAt := strings.Index(routerSource, "middleware.ModelHeaderForwarding(params.Config.ModelHeaders)")
	if forwardingAt < 0 {
		t.Fatalf("ModelHeaderForwarding middleware is not registered in NewRouter")
	}

	publicRoutes := map[string]string{
		"IM callbacks": "RegisterIMRoutes(r, params.IMHandler)",
		"embed routes": "RegisterEmbedPublicRoutes(r, params.EmbedChannelHandler",
	}
	for name, marker := range publicRoutes {
		routeAt := strings.Index(routerSource, marker)
		if routeAt < 0 {
			t.Fatalf("%s marker %q not found", name, marker)
		}
		if forwardingAt > routeAt {
			t.Fatalf("ModelHeaderForwarding is registered after %s; public model calls would miss forwarded headers", name)
		}
	}
}
