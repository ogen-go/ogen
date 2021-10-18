package uri

import (
	"net/url"
	"sync"
)

var urlPool = sync.Pool{
	New: func() interface{} {
		return new(url.URL)
	},
}

// Acquire acquires new url.URL from pool.
func Acquire() *url.URL {
	return urlPool.Get().(*url.URL)
}

// Put puts url.URL to pool.
func Put(u *url.URL) {
	urlPool.Put(u)
}

// Clone clones u and returns cloned url that you can modify.
//
// Call Put for performance.
func Clone(u *url.URL) *url.URL {
	target := Acquire()

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
