package tests

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/generated/oapi"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/handler"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/server"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestEveryOpenAPIOperationIsRegistered(t *testing.T) {
	signer, err := security.NewVisitorSigner([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	application := server.New(server.Dependencies{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), API: handler.New(handler.Dependencies{}), VisitorSigner: signer, RefreshCookieName: "refresh", VisitorCookieName: "visitor", VisitorSecret: []byte("0123456789abcdef0123456789abcdef"), AllowedOrigins: []string{"http://localhost:3000"}})
	registered := make(map[string]struct{})
	for _, route := range application.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}
	specification, err := oapi.GetSwagger()
	require.NoError(t, err)
	for path, item := range specification.Paths.Map() {
		for method, operation := range operations(item) {
			if operation == nil {
				continue
			}
			echoPath := "/api/v1" + strings.ReplaceAll(strings.ReplaceAll(path, "{", ":"), "}", "")
			_, ok := registered[strings.ToUpper(method)+" "+echoPath]
			require.Truef(t, ok, "operation %s is not registered at %s %s", operation.OperationID, method, echoPath)
		}
	}
}

func operations(item *openapi3.PathItem) map[string]*openapi3.Operation {
	return map[string]*openapi3.Operation{"get": item.Get, "post": item.Post, "put": item.Put, "delete": item.Delete, "patch": item.Patch, "head": item.Head, "options": item.Options, "trace": item.Trace}
}
