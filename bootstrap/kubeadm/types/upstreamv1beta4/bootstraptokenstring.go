/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package upstreamv1beta4

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	bootstrapapi "k8s.io/cluster-bootstrap/token/api"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
)

// BootstrapTokenString is a token of the format abcdef.abcdef0123456789 that is used
// for both validation of the practically of the API server from a joining node's point
// of view and as an authentication method for the node in the bootstrap phase of
// "kubeadm join". This token is and should be short-lived.
type BootstrapTokenString struct {
	ID     string `json:"-"`
	Secret string `json:"-" datapolicy:"token"`
}

const (
	// When a token is matched with 'BootstrapTokenPattern', the size of validated substrings returned by
	// regexp functions which contains 'Submatch' in their names will be 3.
	// Submatch 0 is the match of the entire expression, submatch 1 is
	// the match of the first parenthesized subexpression, and so on.
	// e.g.:
	// result := bootstraputil.BootstrapTokenRegexp.FindStringSubmatch("abcdef.1234567890123456")
	// result == []string{"abcdef.1234567890123456","abcdef","1234567890123456"}
	// len(result) == 3.
	validatedSubstringsSize = 3
)

// MarshalJSON implements the json.Marshaler interface.
func (bts BootstrapTokenString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", bts.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (bts *BootstrapTokenString) UnmarshalJSON(b []byte) error {
	// If the token is represented as "", just return quickly without an error
	if len(b) == 0 {
		return nil
	}

	// Remove unnecessary " characters coming from the JSON parser
	token := strings.ReplaceAll(string(b), `"`, ``)
	// Convert the string Token to a BootstrapTokenString object
	newbts, err := NewBootstrapTokenString(token)
	if err != nil {
		return err
	}
	bts.ID = newbts.ID
	bts.Secret = newbts.Secret
	return nil
}

// String returns the string representation of the BootstrapTokenString.
func (bts BootstrapTokenString) String() string {
	if bts.ID != "" && bts.Secret != "" {
		return bootstraputil.TokenFromIDAndSecret(bts.ID, bts.Secret)
	}
	return ""
}

// NewBootstrapTokenString converts the given Bootstrap Token as a string
// to the BootstrapTokenString object used for serialization/deserialization
// and internal usage. It also automatically validates that the given token
// is of the right format.
func NewBootstrapTokenString(token string) (*BootstrapTokenString, error) {
	substrs := bootstraputil.BootstrapTokenRegexp.FindStringSubmatch(token)
	if len(substrs) != validatedSubstringsSize {
		return nil, errors.Errorf("the bootstrap token %q was not of the form %q", token, bootstrapapi.BootstrapTokenPattern)
	}

	return &BootstrapTokenString{ID: substrs[1], Secret: substrs[2]}, nil
}

// NewBootstrapTokenStringFromIDAndSecret is a wrapper around NewBootstrapTokenString
// that allows the caller to specify the ID and Secret separately.
func NewBootstrapTokenStringFromIDAndSecret(id, secret string) (*BootstrapTokenString, error) {
	return NewBootstrapTokenString(bootstraputil.TokenFromIDAndSecret(id, secret))
}
