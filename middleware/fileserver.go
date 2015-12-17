package middleware

import (
	"net/http"
	"os"
	"path"
	"strings"
)

// This file contains a standard way for Caddy middleware
// to load files from the file system given a request
// URI and path to site root. Other middleware that load
// files should use these facilities.

// FileServer implements a production-ready file server
// and is the 'default' handler for all requests to Caddy.
// It simply loads and serves the URI requested. If Caddy is
// run without any extra configuration/directives, this is the
// only middleware handler that runs. It is not in its own
// folder like most other middleware handlers because it does
// not require a directive. It is a special case.
//
// FileServer is adapted from the one in net/http by
// the Go authors. Significant modifications have been made.
//
// Original license:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func FileServer(root http.FileSystem, hide *Hide) Handler {
	return &fileHandler{root: root, hide: hide}
}

type fileHandler struct {
	root http.FileSystem
	hide *Hide // list of files to treat as "Not Found"
}

func (fh *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	return fh.serveFile(w, r, path.Clean(upath))
}

// serveFile writes the specified file to the HTTP response.
// name is '/'-separated, not filepath.Separator.
func (fh *fileHandler) serveFile(w http.ResponseWriter, r *http.Request, name string) (int, error) {

	// If the file is supposed to be hidden, return a 404
	// (TODO: If the slice gets large, a set may be faster)
	if fh.hide.IsMatchPath(name) {
		return http.StatusNotFound, nil
	}

	f, err := fh.root.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return http.StatusNotFound, nil
		} else if os.IsPermission(err) {
			return http.StatusForbidden, err
		}
		// Likely the server is under load and ran out of file descriptors
		w.Header().Set("Retry-After", "5") // TODO: 5 seconds enough delay? Or too much?
		return http.StatusServiceUnavailable, err
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			return http.StatusNotFound, nil
		} else if os.IsPermission(err) {
			return http.StatusForbidden, err
		}
		// Return a different status code than above so as to distinguish these cases
		return http.StatusInternalServerError, err
	}

	// redirect to canonical path
	url := r.URL.Path
	if d.IsDir() {
		// Ensure / at end of directory url
		if url[len(url)-1] != '/' {
			redirect(w, r, path.Base(url)+"/")
			return http.StatusMovedPermanently, nil
		}
	} else {
		// Ensure no / at end of file url
		if url[len(url)-1] == '/' {
			redirect(w, r, "../"+path.Base(url))
			return http.StatusMovedPermanently, nil
		}
	}

	// use contents of an index file, if present, for directory
	if d.IsDir() {
		for _, indexPage := range IndexPages {
			index := strings.TrimSuffix(name, "/") + "/" + indexPage
			ff, err := fh.root.Open(index)
			if err == nil {
				defer ff.Close()
				dd, err := ff.Stat()
				if err == nil {
					name = index
					d = dd
					f = ff
					break
				}
			}
		}
	}

	// Still a directory? (we didn't find an index file)
	// Return 404 to hide the fact that the folder exists
	if d.IsDir() {
		return http.StatusNotFound, nil
	}

	// Note: Errors generated by ServeContent are written immediately
	// to the response. This usually only happens if seeking fails (rare).
	http.ServeContent(w, r, d.Name(), d.ModTime(), f)

	return http.StatusOK, nil
}

// redirect is taken from http.localRedirect of the std lib. It
// sends an HTTP redirect to the client but will preserve the
// query string for the new path.
func redirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	http.Redirect(w, r, newPath, http.StatusMovedPermanently)
}

// IndexPages is a list of pages that may be understood as
// the "index" files to directories.
var IndexPages = []string{
	"index.html",
	"index.htm",
	"index.txt",
	"default.html",
	"default.htm",
	"default.txt",
}


// Config represent a mime config.
type HideConfig struct {
	MatchCase bool
	Prefix []string
	Suffix []string
	Name []string
	Path []string
}

type Hide struct {
	Next Handler
	Configs []HideConfig
}


// Case-insensitive file systems may have loaded "CaddyFile" when
// we think we got "Caddyfile", which poses a security risk if we
// aren't careful here: case-insensitive comparison is required!
// TODO: This matches file NAME only, regardless of path. In other
// words, trying to serve another file with the same name as the
// active config file will result in a 404 when it shouldn't.
func (hide *Hide) IsMatchName(value string) (isMatch bool) {
	vlen := len(value)
	// TODO this does not check the case sensitive flag yet
	for _, config := range hide.Configs {
		for _, prefix := range config.Prefix {
			n := len(prefix)
			if (vlen < n) {
				n = vlen
			}
			if strings.EqualFold(value[0:n], prefix)  {
				return true
			}
		}

		for _, suffix := range config.Suffix {
			n := len(suffix)
			if (vlen < n) {
				n = vlen
			}
			if strings.EqualFold(value[vlen - n:], suffix)  {
				return true
			}
		}

		for _, name := range config.Name {
			if strings.EqualFold(value, name) {
				return true
			}
		}
	}

	return false
}

func (hide *Hide) IsMatchPath(value string) (isMatch bool) {
	if hide == nil {
		return false
	}

	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if hide.IsMatchName(part) {
			return true
		}
	}

	return false
}

// ServeHTTP implements the middleware.Handler interface.
func (e Hide) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	return e.Next.ServeHTTP(w, r)
}