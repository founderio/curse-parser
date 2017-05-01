/*
Copyright 2017 Oliver Kahrmann

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
  or implied. See the License for the specific language governing
  permissions and limitations under the License.
*/

package curse

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gopkg.in/xmlpath.v2"
)

var client *http.Client = &http.Client{}

// FetchPage performs a simple http get using a custom user agent.
func FetchPage(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating http request: %s", err.Error())
	}
	req.Header.Set("User-Agent", "Go-http-client/1.1 (compatible; curse-parser)")

	return client.Do(req)
}

// Instance for internal use.
var pathCache = NewXpathCache()

// XpathCache is a wrapper for the xmlpath package.
// The wrapper functions cache the compiled XPaths instead of recompiling every time.
// The cached instances are kept in this struct. Create a new instance with NewXpathCache().
// An internal instance of this will be used for all parsing operations.
type XpathCache struct {
	paths map[string]*xmlpath.Path
}

// NewXpathCache creates a new, empty XpathCache instance.
func NewXpathCache() *XpathCache {
	return &XpathCache{
		paths: make(map[string]*xmlpath.Path),
	}
}

// GetCompiledPath returns a compiled xpath.
// If the given xpath is already in the cache, the cached instance is returned.
// Otherwise it is compiled (using xmlpath.MustCompile()) and put into cache.
// Panics on compile errors.
func (cache *XpathCache) GetCompiledPath(path string) *xmlpath.Path {
	p, ok := cache.paths[path]
	if !ok {
		p = xmlpath.MustCompile(path)
		cache.AddToCache(path, p)
	}
	return p
}

// Clear the cache of compiled XPaths.
func (cache *XpathCache) Clear() {
	cache.paths = make(map[string]*xmlpath.Path)
}

// AddToCache adds the given compiled path to the cache, identified by the given uncompiled version.
// WARNING: With this, you can thoroughly confuse the cache. This function cannot validate if the compiled version
// is actually matching the uncompiled one!
func (cache *XpathCache) AddToCache(uncompiled string, compiled *xmlpath.Path) {
	cache.paths[uncompiled] = compiled
}

// Wrapper around path.String(context). Trims space.
// The given xpath is automatically compiled or pulled from cache.
func (cache *XpathCache) String(context *xmlpath.Node, path string) (string, bool) {
	p := cache.GetCompiledPath(path)
	s, ok := p.String(context)
	if !ok {
		return s, ok
	}
	return strings.TrimSpace(s), true
}

// Iter is a wrapper around path.Iter(context).
// The given xpath is automatically compiled or pulled from cache.
func (cache *XpathCache) Iter(context *xmlpath.Node, path string) *xmlpath.Iter {
	p := cache.GetCompiledPath(path)
	return p.Iter(context)
}

// Node is a wrapper around path.Iter(context).
// The given xpath is automatically compiled or pulled from cache.
// If something is found, the first node & true is returned.
// If not, nil & false is returned.
func (cache *XpathCache) Node(context *xmlpath.Node, path string) (*xmlpath.Node, bool) {
	iter := cache.Iter(context, path)
	if !iter.Next() {
		return nil, false
	}
	return iter.Node(), true
}

// URL is a wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an URL.
func (cache *XpathCache) URL(context *xmlpath.Node, path string) (*url.URL, error) {
	urlString, ok := cache.String(context, path)
	if !ok {
		return nil, errors.New("node not found")
	}
	return ParseURL(urlString)
}

// ParseURL attempts to parse the given string into a URL.
// Adds the https url scheme if no scheme is missing
// (link urls may be specified in schemeless format "//www.curseforge.com/...")
func ParseURL(urlString string) (*url.URL, error) {
	// Parse to url
	url, err := url.Parse(urlString)
	if err != nil {
		return url, err
	}
	// URL may be specified in schemeless format "//www.curseforge.com/..."
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	return url, nil
}

// URLWithBase is a wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an URL and resolved using 'base'.
// i.e., URLs like
// /members/founderio
// with a base of https://mods.curse.com/mc-mods/minecraft/238424-taam
// will be resolved to https://mods.curse.com/members/founderio.
// Absolute URLs will be returned as-is.
func (cache *XpathCache) URLWithBase(context *xmlpath.Node, path, base string) (*url.URL, error) {
	// Parse the base URL
	url, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return url, err
	}
	// Get value, parse, resolve & fix
	return cache.URLWithBaseURL(context, path, url)
}

// URLWithBaseURL is a wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an URL and resolved using 'base'.
// i.e., URLs like
// /members/founderio
// with a base of https://mods.curse.com/mc-mods/minecraft/238424-taam
// will be resolved to https://mods.curse.com/members/founderio.
// Absolute URLs will be returned as-is.
func (cache *XpathCache) URLWithBaseURL(context *xmlpath.Node, path string, base *url.URL) (*url.URL, error) {
	// Get URL value
	urlString, ok := cache.String(context, path)
	if !ok {
		return nil, errors.New("node not found")
	}
	// Parse to url
	parsedURL, err := url.Parse(strings.TrimSpace(urlString))
	if err != nil {
		return parsedURL, err
	}
	// Resolve & fix
	parsedURL = base.ResolveReference(parsedURL)
	// URL may be specified in schemeless format "//www.curseforge.com/..."
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}
	return parsedURL, nil
}

// UInt is a wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an uint64, base 10. Commas (decimal separator) are stripped before parsing.
func (cache *XpathCache) UInt(context *xmlpath.Node, path string) (uint64, error) {
	parseString, ok := cache.String(context, path)
	if !ok {
		return 0, errors.New("node not found")
	}
	return ParseUInt(parseString)
}

// ParseUInt attempts to parse the given string to an uint64.
// "-" is treated as 0. Commas are removed from the input string.
// (English number format is assumed!)
func ParseUInt(parseString string) (uint64, error) {
	// Extra failsafe for external calls, but not required internally
	str := strings.TrimSpace(parseString)
	// Download counts of '0' are represented as '-'
	if str == "-" {
		return 0, nil
	}
	// No decimal separators please..
	str = strings.Replace(parseString, ",", "", -1)
	return strconv.ParseUint(str, 10, 64)
}

// Int is a wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an int64, base 10. Commas (decimal separator) are stripped before parsing.
func (cache *XpathCache) Int(context *xmlpath.Node, path string) (int64, error) {
	parseString, ok := cache.String(context, path)
	if !ok {
		return 0, errors.New("node not found")
	}
	return ParseInt(parseString)
}

// ParseInt attempts to parse the given string to an int64.
// "-" is treated as 0. Commas are removed from the input string.
// (English number format is assumed!)
func ParseInt(parseString string) (int64, error) {
	// Extra failsafe for external calls, but not required internally
	str := strings.TrimSpace(parseString)
	// Download counts of '0' are represented as '-'
	if str == "-" {
		return 0, nil
	}
	// No decimal separators please..
	str = strings.Replace(parseString, ",", "", -1)
	return strconv.ParseInt(str, 10, 64)
}

// UnixTimestamp is a wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an int64, base 10, and interpreted as a unix time stamp.
// Commas (decimal separator) are stripped before parsing.
// The time.Time returned will be set to UTC. If there is a parsing error, time.Unix(0, 0).UTC() is returned.
func (cache *XpathCache) UnixTimestamp(context *xmlpath.Node, path string) (time.Time, error) {
	ts, err := cache.Int(context, path)
	if err != nil {
		return time.Unix(0, 0).UTC(), err
	}

	return time.Unix(ts, 0).UTC(), nil
}
