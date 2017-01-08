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