// Package introspection provides GraphQL schema introspection capabilities.
package introspection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a GraphQL introspection client.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a new introspection client with an HTTP client.
// The httpClient should already have authentication configured (e.g., from goauth).
func NewClient(endpoint string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		endpoint:   endpoint,
		httpClient: httpClient,
	}
}

// Introspect performs a GraphQL introspection query and returns the result.
func (c *Client) Introspect(ctx context.Context) (*Result, error) {
	reqBody := map[string]any{
		"query": IntrospectionQuery,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var graphqlResp struct {
		Data   *IntrospectionData `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors,omitempty"`
	}

	if err := json.Unmarshal(respBody, &graphqlResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphqlResp.Errors[0].Message)
	}

	if graphqlResp.Data == nil {
		return nil, fmt.Errorf("no data in introspection response")
	}

	return &Result{
		Schema:   graphqlResp.Data.Schema,
		Endpoint: c.endpoint,
	}, nil
}

// IntrospectionQuery is the standard GraphQL introspection query.
const IntrospectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType {
      name
    }
    mutationType {
      name
    }
    subscriptionType {
      name
    }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type {
    ...TypeRef
  }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`
