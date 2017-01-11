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

type CurseForgeSections uint8

const (
	// Header information is always parsed from the first page loaded.
	// Pass only this to load the overview page but only parse the header info.
	CFSectionHeader CurseForgeSections = 0
	// The overview page.
	// Parses the "About this Project" section & other values from the sidebar.
	// Does not include description or comments.
	// Also does not include recent files. Use the option CFOptionOverviewRecentFiles to include them.
	CFSectionOverview = 1
	// The files page.
	// Parses all files of the file page. Multiple pages will be requested sequentially.
	// (always, even when using parallel=true)
	// Use the option CFOptionFilesNoPagination to load only the first page.
	CFSectionFiles  = 2
	CFSectionImages = 4
	// Reserved should we ever want to parse issues on the internal issue tracker.
	_ = 8
)

// Returns true if this sections-selection has the bit for sec set.
func (s CurseForgeSections) Has(sec CurseForgeSections) bool {
	if sec == CFSectionHeader {
		return true
	}
	return (s & sec) != 0
}

type CurseForgeOptions uint8

const (
	CFOptionNone CurseForgeOptions = 0
	// When parsing the overview page, also parse the recent files.
	// They will end up in the same location as when parsing the files page, so files will contain duplicate if you
	// also load the files page when using this. They will also be missing the game version tag, as that is not easily
	// accessible (or even at all for some curseforge sites).
	CFOptionOverviewRecentFiles = 1
	// When parsing the files page
	CFOptionFilesNoPagination = 2
)

// Returns true if this option has the bit for opt set.
func (o CurseForgeOptions) Has(opt CurseForgeOptions) bool {
	if opt == CFOptionNone {
		return true
	}
	return (o & opt) != 0
}

type Author struct {
	Name     string
	Role     string
	URL      *url.URL
	ImageURL *url.URL
}

type Image struct {
	URL          *url.URL
	ThumbnailURL *url.URL
}

type File struct {
	Name               string
	URL                *url.URL
	DirectURL          *url.URL
	ReleaseType        string
	GameVersion        string
	Downloads          uint64
	Date               time.Time
	// The size info as printed on the page, unparsed
	SizeInfo    string
	HasAdditionalFiles bool
}

type Category struct {
	Name     string
	URL      *url.URL
	ImageURL *url.URL
}

type Dependency struct {
	Name     string
	URL      *url.URL
	ImageURL *url.URL
}

type ModsDotCurseDotCom struct {
	Title        string
	DontationURL *url.URL

	Likes     uint64
	Favorites uint64

	Authors    []Author
	Categories []Category
	License    string

	CurseforgeURL *url.URL

	Game    string
	GameURL *url.URL

	AvgDownloads          uint64
	AvgDownloadsTimeframe string
	TotalDownloads        uint64

	Updated time.Time
	Created time.Time

	Screenshots []Image
	Downloads   []File
}

type CurseforgeDotCom struct {
	OverviewURL     *url.URL
	FilesURL        *url.URL
	ImagesURL       *url.URL
	DependenciesURL *url.URL
	DependentsURL   *url.URL

	CurseURL         *url.URL
	ReportProjectURL *url.URL
	IssuesURL        *url.URL
	WikiURL          *url.URL
	SourceURL        *url.URL

	Title               string
	ProjectURL          *url.URL
	DontationURL        *url.URL
	ImageURL            *url.URL
	ImageThumbnailURL   *url.URL
	RootGameCategory    string
	RootGameCategoryURL *url.URL
	License             string
	LicenseURL          *url.URL
	Game                string
	GameURL             *url.URL

	//AvgDownloads          uint64
	//AvgDownloadsTimeframe string
	TotalDownloads uint64

	Created time.Time
	Updated time.Time

	//Likes     uint64
	//Favorites uint64

	Authors    []Author
	Categories []Category

	Screenshots []Image
	Downloads   []File
}
