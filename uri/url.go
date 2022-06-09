package uri

import (
	"net/url"
)

// Clone clones u and returns cloned url that you can modify.
func Clone(u *url.URL) *url.URL {
	target := new(url.URL)

	target.Path = u.Path
	target.RawFragment = u.RawFragment
	target.RawPath = u.RawPath
	target.Scheme = u.Scheme
	target.Host = u.Host
	target.ForceQuery = u.ForceQuery
	target.User = u.User

	return target
}
