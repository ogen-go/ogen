// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpcookie

import (
	"strings"

	"golang.org/x/net/http/httpguts"
)

// Copied from https://github.com/golang/go/blob/bb8f9a6ae66d742cb67b4ad444179905a537de00/src/net/http/cookie.go#L463

// IsCookieNameValid returns true, if cookie name is invalid.
func IsCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}
