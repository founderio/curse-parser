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

// Simple http get download using a custom user agent.
func FetchPage(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating http request: %s", err.Error())
	}
	req.Header.Set("User-Agent", "Go-http-client/1.1 (compatible; curse-parser)")

	return client.Do(req)
}

// A wrapper for the xmlpath package.
// The wrapper functions cache the compiled XPaths instead of recompiling every time.
// The cached instances are kept in this struct. Create a new instance with NewXpathCache().
type XpathCache struct {
	paths map[string]*xmlpath.Path
}

// Create a new, empty XpathCache instance.
func NewXpathCache() *XpathCache {
	return &XpathCache{
		paths: make(map[string]*xmlpath.Path),
	}
}

// Get or compile an XPath.
// If the given xpath is already in the cache, the cached instance is returned.
// Otherwise it is compiled (using xmlpath.MustCompile()) and put into cache.
func (cache *XpathCache) GetCompiledPath(path string) *xmlpath.Path {
	p, ok := cache.paths[path]
	if !ok {
		p = xmlpath.MustCompile(path)
		cache.AddToCache(path, p)
	}
	return p
}

// Clears the cache of compiled XPaths.
func (cache *XpathCache) Clear() {
	cache.paths = make(map[string]*xmlpath.Path)
}

// Adds the given compiled path to the cache, identified by the given uncompiled version.
// WARNING: With this, you can thoroughly confuse the cache. This function cannot validate if the compiled version
// is actually matching the uncompiled one!
func (cache *XpathCache) AddToCache(uncompiled string, compiled *xmlpath.Path) {
	cache.paths[uncompiled] = compiled
}

// Wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
func (cache *XpathCache) String(context *xmlpath.Node, path string) (string, bool) {
	p := cache.GetCompiledPath(path)
	return p.String(context)
}

// Wrapper around path.Iter(context).
// The given xpath is automatically compiled or pulled from cache.
func (cache *XpathCache) Iter(context *xmlpath.Node, path string) *xmlpath.Iter {
	p := cache.GetCompiledPath(path)
	return p.Iter(context)
}

// Wrapper around path.Iter(context).
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

// Wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an URL.
func (cache *XpathCache) URL(context *xmlpath.Node, path string) (*url.URL, error) {
	urlString, ok := cache.String(context, path)
	if !ok {
		return nil, errors.New("node not found")
	}
	// Parse to url
	url, err := url.Parse(strings.TrimSpace(urlString))
	if err != nil {
		return url, err
	}
	// URL may be specified in schemeless format "//www.curseforge.com/..."
	if url.Scheme == "" {
		url.Scheme = "https"
	}
	return url, nil
}

// Wrapper around path.String(context).
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

// Wrapper around path.String(context).
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
	parsedUrl, err := url.Parse(strings.TrimSpace(urlString))
	if err != nil {
		return parsedUrl, err
	}
	// Resolve & fix
	parsedUrl = base.ResolveReference(parsedUrl)
	// URL may be specified in schemeless format "//www.curseforge.com/..."
	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = "https"
	}
	return parsedUrl, nil
}

// Wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an uint64, base 10. Commas (decimal separator) are stripped before parsing.
func (cache *XpathCache) UInt(context *xmlpath.Node, path string) (uint64, error) {
	parseString, ok := cache.String(context, path)
	if !ok {
		return 0, errors.New("node not found")
	}
	return strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
}

// Wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an int64, base 10. Commas (decimal separator) are stripped before parsing.
func (cache *XpathCache) Int(context *xmlpath.Node, path string) (int64, error) {
	parseString, ok := cache.String(context, path)
	if !ok {
		return 0, errors.New("node not found")
	}
	return strconv.ParseInt(strings.Replace(parseString, ",", "", -1), 10, 64)
}

// Wrapper around path.String(context).
// The given xpath is automatically compiled or pulled from cache.
// The returned value is parsed to an int64, base 10, and interpreted as a unix time stamp.
// Commas (decimal separator) are stripped before parsing.
// The time.Time returned will be set to UTC.
func (cache *XpathCache) UnixTimestamp(context *xmlpath.Node, path string) (time.Time, error) {
	ts, err := cache.Int(context, path)
	if err != nil {
		return time.Now(), err
	}

	return time.Unix(ts, 0).UTC(), nil
}
