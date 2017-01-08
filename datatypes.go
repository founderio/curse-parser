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
	"net/url"
	"time"
)

type Author struct {
	Name string
	Role string
	URL  *url.URL
}

type Image struct {
	URL          *url.URL
	ThumbnailURL *url.URL
}

type Download struct {
	Name        string
	URL         *url.URL
	ReleaseType string
	GameVersion string
	Downloads   uint64
	Date        time.Time
}

type Category struct {
	Name     string
	URL      *url.URL
	ImageURL *url.URL
}

type ModsDotCurseDotCom struct {
	Title        string
	DontationURL *url.URL

	Likes     uint64
	Favorites uint64

	Authors       []Author
	Categories    []Category
	License       string
	CurseforgeURL *url.URL

	Game    string
	GameURL *url.URL

	AvgDownloads          uint64
	AvgDownloadsTimeframe string
	TotalDownloads        uint64

	Updated time.Time
	Created time.Time

	Screenshots []Image
	Downloads   []Download
}
