package gen

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/jsonschema"
)

// RemoteOptions is remote reference resolver options.
type RemoteOptions struct {
	// HTTPClient sets http client to use. Defaults to http.DefaultClient.
	HTTPClient *http.Client
	// ReadFile sets function for reading files from fs. Defaults to os.ReadFile.
	ReadFile func(p string) ([]byte, error)
}

func (r *RemoteOptions) setDefaults() {
	if r.HTTPClient == nil {
		r.HTTPClient = http.DefaultClient
	}
	if r.ReadFile == nil {
		r.ReadFile = os.ReadFile
	}
}

var _ jsonschema.ExternalResolver = externalResolver{}

type externalResolver struct {
	client   *http.Client
	readFile func(p string) ([]byte, error)
}

func newExternalResolver(opts RemoteOptions) externalResolver {
	opts.setDefaults()

	return externalResolver{
		client:   opts.HTTPClient,
		readFile: opts.ReadFile,
	}
}

func (e externalResolver) httpGet(ctx context.Context, u *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}
	if pass, ok := u.User.Password(); ok && u.User != nil {
		req.SetBasicAuth(u.User.Username(), pass)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do")
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if code := resp.StatusCode; code >= 299 {
		text := http.StatusText(code)
		return nil, errors.Errorf("bad HTTP code %d (%s)", code, text)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read data")
	}

	return data, nil
}

func (e externalResolver) Get(ctx context.Context, loc string) ([]byte, error) {
	u, err := url.Parse(loc)
	if err != nil {
		return nil, err
	}

	var (
		data   []byte
		scheme = u.Scheme
	)
	switch scheme {
	case "http", "https":
		data, err = e.httpGet(ctx, u)
	case "file", "":
		data, err = e.readFile(u.Path)
	default:
		return nil, errors.Errorf("unsupported scheme %q", scheme)
	}
	if err != nil {
		if scheme == "" {
			scheme = "file"
		}
		return nil, errors.Wrap(err, scheme)
	}

	return data, nil
}
