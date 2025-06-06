package transport_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/99designs/gqlgen/graphql/handler/testserver"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

func TestGET(t *testing.T) {
	h := testserver.New()
	h.AddTransport(transport.GET{})

	graphqlResponseH := testserver.New()
	graphqlResponseH.AddTransport(transport.GET{UseGrapQLResponseJsonByDefault: true})

	jsonH := testserver.New()
	jsonH.AddTransport(transport.GET{
		ResponseHeaders: map[string][]string{
			"Content-Type": {"application/json"},
		},
	})

	t.Run("success with accept application/json", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query={name}", ``, "application/json", "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"data":{"name":"test"}}`, resp.Body.String())
	})

	t.Run("success with accept is empty with enabling graphql response json", func(t *testing.T) {
		resp := doRequest(graphqlResponseH, "GET", "/graphql?query={name}", ``, "", "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, "application/graphql-response+json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"data":{"name":"test"}}`, resp.Body.String())
	})

	t.Run("success with accept is empty without enabling graphql response json", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query={name}", ``, "", "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"data":{"name":"test"}}`, resp.Body.String())
	})

	t.Run("success with accept application/graphql-response+json", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query={name}", ``, "application/graphql-response+json", "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, "application/graphql-response+json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"data":{"name":"test"}}`, resp.Body.String())
	})

	t.Run("success with wildcard with enabling application/graphql-response+json", func(t *testing.T) {
		resp := doRequest(graphqlResponseH, "GET", "/graphql?query={name}", ``, "*/*", "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, "application/graphql-response+json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"data":{"name":"test"}}`, resp.Body.String())
	})

	t.Run("success with wildcard without enabling application/graphql-response+json", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query={name}", ``, "*/*", "application/json")
		assert.Equal(t, http.StatusOK, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"data":{"name":"test"}}`, resp.Body.String())
	})

	t.Run("has json content-type header", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query={name}", ``, "application/json", "application/json")
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	})

	t.Run("decode failure", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query={name}&variables=notjson", "", "application/json", "application/json")
		assert.Equal(t, http.StatusBadRequest, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"errors":[{"message":"variables could not be decoded"}],"data":null}`, resp.Body.String())
	})

	t.Run("invalid variable", func(t *testing.T) {
		resp := doRequest(h, "GET", `/graphql?query=query($id:Int!){find(id:$id)}&variables={"id":false}`, "", "", "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"errors":[{"message":"cannot use bool as Int","path":["variable","id"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`, resp.Body.String())
	})

	t.Run("invalid variable with json only", func(t *testing.T) {
		resp := doRequest(jsonH, "GET", `/graphql?query=query($id:Int!){find(id:$id)}&variables={"id":false}`, "", "", "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"errors":[{"message":"cannot use bool as Int","path":["variable","id"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}],"data":null}`, resp.Body.String())
	})

	t.Run("parse failure", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query=!", "", "", "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"errors":[{"message":"Unexpected !","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_PARSE_FAILED"}}],"data":null}`, resp.Body.String())
	})

	t.Run("parse failure with json only", func(t *testing.T) {
		resp := doRequest(jsonH, "GET", "/graphql?query=!", "", "", "application/json")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code, resp.Body.String())
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"errors":[{"message":"Unexpected !","locations":[{"line":1,"column":1}],"extensions":{"code":"GRAPHQL_PARSE_FAILED"}}],"data":null}`, resp.Body.String())
	})

	t.Run("no mutations", func(t *testing.T) {
		resp := doRequest(h, "GET", "/graphql?query=mutation{name}", "", "", "application/json")
		assert.Equal(t, http.StatusNotAcceptable, resp.Code, resp.Body.String())
		assert.JSONEq(t, `{"errors":[{"message":"GET requests only allow query operations"}],"data":null}`, resp.Body.String())
	})
}
