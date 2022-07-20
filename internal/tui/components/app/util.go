package app

import (
	"github.com/hashicorp/nomad/api"
	"strings"
	"sync"
)

var (
	updateID    int
	updateIDMtx sync.Mutex
)

func nextUpdateID() int {
	updateIDMtx.Lock()
	defer updateIDMtx.Unlock()
	updateID++
	return updateID
}

func (c Config) client() (*api.Client, error) {
	config := &api.Config{
		Address:   c.URL,
		SecretID:  c.Token,
		Region:    c.Region,
		Namespace: c.Namespace,
		TLSConfig: &api.TLSConfig{
			CACert:        c.TLS.CACert,
			CAPath:        c.TLS.CAPath,
			ClientCert:    c.TLS.ClientCert,
			ClientKey:     c.TLS.ClientKey,
			TLSServerName: c.TLS.ServerName,
			Insecure:      c.TLS.SkipVerify,
		},
	}

	if auth := c.HTTPAuth; auth != "" {
		var username, password string
		if strings.Contains(auth, ":") {
			split := strings.SplitN(auth, ":", 2)
			username = split[0]
			password = split[1]
		} else {
			username = auth
		}

		config.HttpAuth = &api.HttpBasicAuth{
			Username: username,
			Password: password,
		}
	}

	return api.NewClient(config)
}
