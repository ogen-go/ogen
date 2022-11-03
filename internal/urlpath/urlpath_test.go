// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package urlpath

import (
	"net/url"
	"runtime"
	"testing"
)

func TestURLToFilePath(t *testing.T) {
	for _, tc := range urlTests() {
		if tc.url == "" {
			continue
		}
		tc := tc

		t.Run(tc.url, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("url.Parse(%q): %v", tc.url, err)
			}

			path, err := URLToFilePath(u)
			if err != nil {
				if err.Error() == tc.wantErr {
					return
				}
				if tc.wantErr == "" {
					t.Fatalf("urlToFilePath(%v): %v; want <nil>", u, err)
				} else {
					t.Fatalf("urlToFilePath(%v): %v; want %s", u, err, tc.wantErr)
				}
			}

			if path != tc.filePath || tc.wantErr != "" {
				t.Fatalf("urlToFilePath(%v) = %q, <nil>; want %q, %s", u, path, tc.filePath, tc.wantErr)
			}
		})
	}
}

func TestURLFromFilePath(t *testing.T) {
	for _, tc := range urlTests() {
		if tc.filePath == "" {
			continue
		}
		tc := tc

		t.Run(tc.filePath, func(t *testing.T) {
			u, err := URLFromFilePath(tc.filePath)
			if err != nil {
				if err.Error() == tc.wantErr {
					return
				}
				if tc.wantErr == "" {
					t.Fatalf("urlFromFilePath(%v): %v; want <nil>", tc.filePath, err)
				} else {
					t.Fatalf("urlFromFilePath(%v): %v; want %s", tc.filePath, err, tc.wantErr)
				}
			}

			if tc.wantErr != "" {
				t.Fatalf("urlFromFilePath(%v) = <nil>; want error: %s", tc.filePath, tc.wantErr)
			}

			wantURL := tc.url
			if tc.canonicalURL != "" {
				wantURL = tc.canonicalURL
			}
			if u.String() != wantURL {
				t.Errorf("urlFromFilePath(%v) = %v; want %s", tc.filePath, u, wantURL)
			}
		})
	}
}

func urlTests() []urlTest {
	if runtime.GOOS == "windows" {
		return urlTestsWindows
	}
	return urlTestsOthers
}

type urlTest struct {
	url          string
	filePath     string
	canonicalURL string // If empty, assume equal to url.
	wantErr      string
}

var urlTestsOthers = []urlTest{
	// Examples from RFC 8089:
	{
		url:      `file:///path/to/file`,
		filePath: `/path/to/file`,
	},
	{
		url:          `file:/path/to/file`,
		filePath:     `/path/to/file`,
		canonicalURL: `file:///path/to/file`,
	},
	{
		url:          `file://localhost/path/to/file`,
		filePath:     `/path/to/file`,
		canonicalURL: `file:///path/to/file`,
	},

	// We reject non-local files.
	{
		url:     `file://host.example.com/path/to/file`,
		wantErr: "file URL specifies non-local host",
	},
}

var urlTestsWindows = []urlTest{
	// Examples from https://blogs.msdn.microsoft.com/ie/2006/12/06/file-uris-in-windows/:

	{
		url:      `file://laptop/My%20Documents/FileSchemeURIs.doc`,
		filePath: `\\laptop\My Documents\FileSchemeURIs.doc`,
	},
	{
		url:      `file:///C:/Documents%20and%20Settings/davris/FileSchemeURIs.doc`,
		filePath: `C:\Documents and Settings\davris\FileSchemeURIs.doc`,
	},
	{
		url:      `file:///D:/Program%20Files/Viewer/startup.htm`,
		filePath: `D:\Program Files\Viewer\startup.htm`,
	},
	{
		url:          `file:///C:/Program%20Files/Music/Web%20Sys/main.html?REQUEST=RADIO`,
		filePath:     `C:\Program Files\Music\Web Sys\main.html`,
		canonicalURL: `file:///C:/Program%20Files/Music/Web%20Sys/main.html`,
	},
	{
		url:      `file://applib/products/a-b/abc_9/4148.920a/media/start.swf`,
		filePath: `\\applib\products\a-b\abc_9\4148.920a\media\start.swf`,
	},
	{
		url:     `file:////applib/products/a%2Db/abc%5F9/4148.920a/media/start.swf`,
		wantErr: "file URL missing drive letter",
	},
	{
		url:     `C:\Program Files\Music\Web Sys\main.html?REQUEST=RADIO`,
		wantErr: "non-file URL",
	},

	// The example "file://D:\Program Files\Viewer\startup.htm" errors out in
	// url.Parse, so we substitute a slash-based path for testing instead.
	{
		url:     `file://D:/Program Files/Viewer/startup.htm`,
		wantErr: "file URL encodes volume in host field: too few slashes?",
	},

	// The blog post discourages the use of non-ASCII characters because they
	// depend on the user's current codepage. However, when we are working with Go
	// strings we assume UTF-8 encoding, and our url package refuses to encode
	// URLs to non-ASCII strings.
	{
		url:          `file:///C:/exampleㄓ.txt`,
		filePath:     `C:\exampleㄓ.txt`,
		canonicalURL: `file:///C:/example%E3%84%93.txt`,
	},
	{
		url:      `file:///C:/example%E3%84%93.txt`,
		filePath: `C:\exampleㄓ.txt`,
	},

	// Examples from RFC 8089:

	// We allow the drive-letter variation from section E.2, because it is
	// simpler to support than not to. However, we do not generate the shorter
	// form in the reverse direction.
	{
		url:          `file:c:/path/to/file`,
		filePath:     `c:\path\to\file`,
		canonicalURL: `file:///c:/path/to/file`,
	},

	// We encode the UNC share name as the authority following section E.3.1,
	// because that is what the Microsoft blog post explicitly recommends.
	{
		url:      `file://host.example.com/Share/path/to/file.txt`,
		filePath: `\\host.example.com\Share\path\to\file.txt`,
	},

	// We decline the four- and five-slash variations from section E.3.2.
	// The paths in these URLs would change meaning under path.Clean.
	{
		url:     `file:////host.example.com/path/to/file`,
		wantErr: "file URL missing drive letter",
	},
	{
		url:     `file://///host.example.com/path/to/file`,
		wantErr: "file URL missing drive letter",
	},
}
