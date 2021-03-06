/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package apiserver

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/auth/authenticator"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/auth/authenticator/bearertoken"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/GoogleCloudPlatform/kubernetes/plugin/pkg/auth/authenticator/password/passwordfile"
	"github.com/GoogleCloudPlatform/kubernetes/plugin/pkg/auth/authenticator/request/basicauth"
	"github.com/GoogleCloudPlatform/kubernetes/plugin/pkg/auth/authenticator/request/union"
	"github.com/GoogleCloudPlatform/kubernetes/plugin/pkg/auth/authenticator/request/x509"
	"github.com/GoogleCloudPlatform/kubernetes/plugin/pkg/auth/authenticator/token/tokenfile"
)

// NewAuthenticator returns an authenticator.Request or an error
func NewAuthenticator(basicAuthFile, clientCAFile, tokenFile string) (authenticator.Request, error) {
	var authenticators []authenticator.Request

	if len(basicAuthFile) > 0 {
		basicAuth, err := newAuthenticatorFromBasicAuthFile(basicAuthFile)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, basicAuth)
	}

	if len(clientCAFile) > 0 {
		certAuth, err := newAuthenticatorFromClientCAFile(clientCAFile)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, certAuth)
	}

	if len(tokenFile) > 0 {
		tokenAuth, err := newAuthenticatorFromTokenFile(tokenFile)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, tokenAuth)
	}

	switch len(authenticators) {
	case 0:
		return nil, nil
	case 1:
		return authenticators[0], nil
	default:
		return union.New(authenticators...), nil
	}
}

// newAuthenticatorFromBasicAuthFile returns an authenticator.Request or an error
func newAuthenticatorFromBasicAuthFile(basicAuthFile string) (authenticator.Request, error) {
	basicAuthenticator, err := passwordfile.NewCSV(basicAuthFile)
	if err != nil {
		return nil, err
	}

	return basicauth.New(basicAuthenticator), nil
}

// newAuthenticatorFromTokenFile returns an authenticator.Request or an error
func newAuthenticatorFromTokenFile(tokenAuthFile string) (authenticator.Request, error) {
	tokenAuthenticator, err := tokenfile.NewCSV(tokenAuthFile)
	if err != nil {
		return nil, err
	}

	return bearertoken.New(tokenAuthenticator), nil
}

// newAuthenticatorFromClientCAFile returns an authenticator.Request or an error
func newAuthenticatorFromClientCAFile(clientCAFile string) (authenticator.Request, error) {
	roots, err := util.CertPoolFromFile(clientCAFile)
	if err != nil {
		return nil, err
	}

	opts := x509.DefaultVerifyOptions()
	opts.Roots = roots

	return x509.New(opts, x509.CommonNameUserConversion), nil
}
