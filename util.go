/*
Copyright $today.year Oliver Kahrmann

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
	"fmt"
	"net/http"
	"net/url"
	"gopkg.in/xmlpath.v2"
	"errors"
	"strings"
	"strconv"
	"time"
)

var client *http.Client = &http.Client{}

func FetchPage(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating http request: %s", err.Error())
	}
	req.Header.Set("User-Agent", "Go-http-client/1.1 (compatible; curse-parser)")

	return client.Do(req)
}

func combineUrl(baseUrl, path string) (*url.URL, error) {
	//TODO: theoretically, parsing once is enough -> instead of parsing once per author... So, pass a *url.URL in here directly
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %s", err.Error())
	}
	u.Path = path
	return u, nil
}

type XpathCache struct {
	paths map[string]*xmlpath.Path
}

func NewXpathCache() *XpathCache {
	return &XpathCache{
		paths: make(map[string]*xmlpath.Path),
	}
}

func (cache *XpathCache) GetCompiledPath(path string) *xmlpath.Path {
	p, ok := cache.paths[path]
	if !ok {
		p = xmlpath.MustCompile(path)
		cache.paths[path] = p
	}
	return p
}

func (cache *XpathCache) String(context *xmlpath.Node, path string) (string, bool) {
	p := cache.GetCompiledPath(path)
	return p.String(context)
}

func (cache *XpathCache) Iter(context *xmlpath.Node, path string) *xmlpath.Iter {
	p := cache.GetCompiledPath(path)
	return p.Iter(context)
}

func (cache *XpathCache) Node(context *xmlpath.Node, path string) (*xmlpath.Node, bool) {
	iter := cache.Iter(context, path)
	if !iter.Next() {
		return nil, false
	}
	return iter.Node(), true
}

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

func (cache *XpathCache) URLWithBase(context *xmlpath.Node, path, base string) (*url.URL, error) {
	// Parse the base URL
	url, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return url, err
	}
	// Get value, parse, resolve & fix
	return cache.URLWithBaseURL(context, path, url)
}

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

func (cache *XpathCache) UInt(context *xmlpath.Node, path string) (uint64, error) {
	parseString, ok := cache.String(context, path)
	if !ok {
		return 0, errors.New("node not found")
	}
	return strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
}

func (cache *XpathCache) Int(context *xmlpath.Node, path string) (int64, error) {
	parseString, ok := cache.String(context, path)
	if !ok {
		return 0, errors.New("node not found")
	}
	return strconv.ParseInt(strings.Replace(parseString, ",", "", -1), 10, 64)
}

func (cache *XpathCache) UnixTimestamp(context *xmlpath.Node, path string) (time.Time, error) {
	ts, err := cache.Int(context, path)
	if err != nil {
		return time.Now(), err
	}

	return time.Unix(ts, 0).UTC(), nil
}