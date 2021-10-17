package http

import (
	"net/url"
	"sync"
)

var urlPool = sync.Pool{
	New: func() interface{} {
		return new(url.URL)
	},
}

// AcquireURL acquires new url.URL from pool.
func AcquireURL() *url.URL {
	return urlPool.Get().(*url.URL)
}

// PutURL puts url.URL to pool.
func PutURL(u *url.URL) {
	urlPool.Put(u)
}

// CloneURL clones u and returns cloned url that you can modify.
//
// Cal PutURL for performance.
func CloneURL(u *url.URL) *url.URL {
	target := AcquireURL()

	target.Path = u.Path
	target.RawFragment = u.RawFragment
	target.RawPath = u.RawPath
	target.Scheme = u.Scheme
	target.Host = u.Host
	target.ForceQuery = u.ForceQuery

	// HACK: Should we copy here?
	target.User = u.User

	return target
}
