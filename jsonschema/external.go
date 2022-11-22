package jsonschema

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/internal/urlpath"
)

// ExternalResolver resolves external links.
type ExternalResolver interface {
	Get(ctx context.Context, loc string) ([]byte, error)
}

var _ ExternalResolver = NoExternal{}

// NoExternal is ExternalResolver that always returns error.
type NoExternal struct{}

// Get implements ExternalResolver.
func (n NoExternal) Get(context.Context, string) ([]byte, error) {
	return nil, errors.New("external references are disabled")
}

// ExternalOptions is external reference resolver options.
type ExternalOptions struct {
	// HTTPClient sets http client to use. Defaults to http.DefaultClient.
	HTTPClient *http.Client

	// ReadFile sets function for reading files from fs. Defaults to os.ReadFile.
	ReadFile func(p string) ([]byte, error)
	// URLToFilePath sets function for converting url to file path. Defaults to urlpath.URLToFilePath.
	URLToFilePath func(u *url.URL) (string, error)

	// Logger sets logger to use. Defaults to zap.NewNop().
	Logger *zap.Logger
}

func (r *ExternalOptions) setDefaults() {
	if r.HTTPClient == nil {
		r.HTTPClient = http.DefaultClient
	}
	if r.ReadFile == nil {
		r.ReadFile = os.ReadFile
	}
	if r.URLToFilePath == nil {
		r.URLToFilePath = urlpath.URLToFilePath
	}
	if r.Logger == nil {
		r.Logger = zap.NewNop()
	}
}

var _ ExternalResolver = externalResolver{}

type externalResolver struct {
	client        *http.Client
	readFile      func(p string) ([]byte, error)
	urlToFilePath func(u *url.URL) (string, error)
	logger        *zap.Logger
}

// NewExternalResolver creates new ExternalResolver.
//
// Currently only http(s) and file schemes are supported.
func NewExternalResolver(opts ExternalOptions) ExternalResolver {
	opts.setDefaults()

	return externalResolver{
		client:        opts.HTTPClient,
		readFile:      opts.ReadFile,
		urlToFilePath: opts.URLToFilePath,
		logger:        opts.Logger,
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

	start := time.Now()
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do")
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	e.logger.Debug("Get",
		zap.String("url", u.Redacted()),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(start)),
	)

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
		var p string
		p, err = e.urlToFilePath(u)
		if err != nil {
			err = errors.Wrap(err, "convert url to file path")
			break
		}
		data, err = e.readFile(p)
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
