// pixeldrain.go
//
// Copyright (c) 2018-2023 Junpei Kawamoto
//
// This software is released under the MIT License.
//
// http://opensource.org/licenses/mit-license.php

package pixeldrain

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/jkawamoto/go-pixeldrain/client"
)

// Default is a default client.
var Default = New(nil, nil)

const (
	// fileBasePath is the base path to file download URLs.
	fileBasePath = "/file"
	// listBasePath is the base path to list URLs.
	listBasePath = "/l"
)

// New creates a new PixelDrain client with the given configurations. formats and cfg can be nil.
func New(formats strfmt.Registry, cfg *client.TransportConfig) *client.PixeldrainAPI {
	if cfg == nil {
		cfg = client.DefaultTransportConfig()
	}

	// create custom transport
	customTransport := new(http.Transport)
	setDefaults(customTransport, http.DefaultTransport)
	customTransport.MaxIdleConns = 8
	customTransport.IdleConnTimeout = 30 * time.Second
	customTransport.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}

	// create runtime client
	transport := runtimeClient.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	transport.Transport = customTransport
	transport.EnableConnectionReuse()

	// create pixeldrain client
	cli := client.New(transport, formats)
	cli.SetTransport(ContentTypeFixer(cli.Transport))

	return cli
}

// DownloadURL returns the URL associated with the given file ID.
func DownloadURL(id string) string {
	u := url.URL{
		Scheme: client.DefaultSchemes[0],
	}
	return u.JoinPath(client.DefaultHost, client.DefaultBasePath, fileBasePath, id).String()
}

// ListURL returns the URL associated with the given list ID.
func ListURL(id string) string {
	u := url.URL{
		Scheme: client.DefaultSchemes[0],
	}
	return u.JoinPath(client.DefaultHost, listBasePath, id).String()
}

// IsDownloadURL returns true if the given url points a file.
func IsDownloadURL(u string) (bool, error) {
	parse, err := url.Parse(u)
	if err != nil {
		return false, err
	}

	prefix, err := url.JoinPath(client.DefaultBasePath, fileBasePath)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(parse.Path, prefix), nil
}

// IsListURL returns true if the given url points a list.
func IsListURL(u string) (bool, error) {
	parse, err := url.Parse(u)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(parse.Path, listBasePath), nil
}

// setDefaults for a from b
//
// a and b should be pointers to the same kind of struct
//
// This copies the public members only from b to a.  This is useful if
// you can't just use a struct copy because it contains a private
// mutex, e.g. as http.Transport.
func setDefaults(a, b any) {
	pt := reflect.TypeOf(a)
	t := pt.Elem()
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()
	for i := range t.NumField() {
		aField := va.Field(i)
		// Set a from b if it is public
		if aField.CanSet() {
			bField := vb.Field(i)
			aField.Set(bField)
		}
	}
}
