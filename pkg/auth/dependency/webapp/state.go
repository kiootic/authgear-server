package webapp

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

type State struct {
	ID              string                 `json:"id"`
	Error           *skyerr.APIError       `json:"error"`
	RedirectURI     string                 `json:"redirect_uri,omitempty"`
	KeepState       bool                   `json:"keep_state,omitempty"`
	GraphInstanceID string                 `json:"graph_instance_id,omitempty"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
}

// Attach attaches s to input.
func (s *State) Attach(input *url.URL) *url.URL {
	u := *input

	q := u.Query()
	q.Set("x_sid", s.ID)

	u.Scheme = ""
	u.Opaque = ""
	u.Host = ""
	u.User = nil
	u.RawQuery = q.Encode()

	return &u
}