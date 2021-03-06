package requests

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Chyroc/phpmyadmin-cli/internal/common"
)

func request(session *Session, method, uri, path string, query, header map[string]string, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		uri = "http://" + uri
	}

	URL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if path != "" {
		URL.Path = path
	}

	if query != nil {
		q := URL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		URL.RawQuery = q.Encode()
	}

	common.Debug("http %s %s with %#v\n", method, URL.String(), body)

	req, err := http.NewRequest(method, URL.String(), body)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	return session.client.Do(req)
}
