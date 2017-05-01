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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"path"

	"gopkg.in/xmlpath.v2"
)

// CurseForgeSections defines which sections (sub-pages) of a curseforge project
// are to be parsed.
// Multiple sections can be added together.
type CurseForgeSections uint8

const (
	// CFSectionHeader enables parsing of the header information present of every page.
	// Header information is always parsed from the first page loaded.
	// Pass only this to load the overview page but only parse the header info.
	CFSectionHeader CurseForgeSections = 0
	// CFSectionOverview enables fetching of the overview page.
	// Parses the "About this Project" section & other values from the sidebar.
	// Does not include description or comments.
	// Also does not include recent files. Use the option CFOptionOverviewRecentFiles to include them.
	CFSectionOverview = 1
	// CFSectionFiles enables fetching of the files page.
	// Parses all files of the file page. Multiple pages will be requested sequentially.
	// (always, even when using parallel=true)
	// Use the option CFOptionFilesNoPagination to load only the first page.
	CFSectionFiles = 2
	// CFSectionImages enables fetching of the images page. Currently not implemented!
	CFSectionImages = 4
	// Reserved should we ever want to parse issues on the internal issue tracker.
	_ = 8
)

// Has is a convenience function for binary operations.
// It returns true if this sections-selection has the bit for sec set.
func (s CurseForgeSections) Has(sec CurseForgeSections) bool {
	if sec == CFSectionHeader {
		return true
	}
	return (s & sec) != 0
}

// CurseForgeOptions allow tweaks to the parsers of some sub-pages.
// See documentation on the single flags for details.
type CurseForgeOptions uint8

const (
	// CFOptionNone defines no specific option. Defaults are used.
	CFOptionNone CurseForgeOptions = 0
	// CFOptionOverviewRecentFiles instructs the overview parser, when parsing
	// the overview page, to also parse the recent files.
	// They will end up in the same location as when parsing the files page,
	// so files will contain duplicate if you also load the files page when
	// using this. They will also be missing the game version tag, as that
	// is not easily accessible (or even at all for some curseforge sites).
	CFOptionOverviewRecentFiles = 1
	// CFOptionFilesNoPagination instructs the files parser to ignore
	// subsequent files pages. Only the first page of files will be parsed.
	CFOptionFilesNoPagination = 2
)

// Has is a convenience function for binary operations.
// It returns true if this option has the bit for opt set.
func (o CurseForgeOptions) Has(opt CurseForgeOptions) bool {
	if opt == CFOptionNone {
		return true
	}
	return (o & opt) != 0
}

// DeriveCurseForgeURLs derives sub-urls like files or images from the
// base projectURL on CurseForge. No HTTP calls are made.
// Example:
// https://minecraft.curseforge.com/projects/taam
// would be derived to a map with this content:
// CFSectionOverview -> https://minecraft.curseforge.com/projects/taam
// CFSectionFiles -> https://minecraft.curseforge.com/projects/taam/files
// CFSectionImages -> https://minecraft.curseforge.com/projects/taam/images
func DeriveCurseForgeURLs(projectURL *url.URL) (map[CurseForgeSections]*url.URL, error) {
	urls := make(map[CurseForgeSections]*url.URL, 5)
	relatives := make(map[CurseForgeSections]string, 5)
	relatives[CFSectionFiles] = "files"
	relatives[CFSectionImages] = "images"

	// Overview does not have a "subfolder"
	urls[CFSectionOverview] = projectURL

	var err error

	for section, rel := range relatives {
		urls[section] = &url.URL{
			Path:     path.Join(projectURL.Path, rel),
			Host:     projectURL.Host,
			Scheme:   projectURL.Scheme,
			User:     projectURL.User,
			RawQuery: projectURL.RawQuery,
		}
		if err != nil {
			return nil, fmt.Errorf("Error deriving url '%s': %s", rel, err.Error())
		}
	}
	return urls, nil
}

// FetchCurseForge fetches and parses mod pages from curseforge.com. The sections to be fetched & parsed can be selected.
// (minecraft.curseforge.com, or feed-the-beast.com, or or or - For a complete list, see curseforge.com).
//
// This function expects to get the project URL (e.g. "https://minecraft.curseforge.com/projects/taam") and will build
// all other required URLs based on the content selection.
//
// Other parameters:
//
// sections: Defines the content pages to be fetched & parsed.
//
// If 0 (CFHeader) is passed only, the overview page
// is loaded and only the header values are parsed & returned.
// Multiple values can be added to select multiple sections,
// e.g. CFSectionFiles | CFSectionImages
// This causes multiple sequential requests to CurseForge.
//
// options: Pass CFOptionNone, or one or more of the other options to tweak
// some behaviour of the content parsers.
//
// Multiple values can be added to define multiple options,
// e.g. CFOptionOverviewRecentFiles | CFOptionFilesNoPagination
func FetchCurseForge(projectURL *url.URL, sections CurseForgeSections, options CurseForgeOptions) (*CurseForge, error) {
	results := new(CurseForge)

	// if the requested section is 0 (CFSectionHeader) we load the overview page, and only parse the header
	if sections == CFSectionHeader {
		var resp *http.Response

		resp, err := FetchPage(projectURL.String())
		if err != nil {
			return nil, err
		}
		err = results.ParseCurseForge(projectURL, resp, true, CFSectionHeader, options)
		if err != nil {
			return nil, err
		}

	} else {

		// Derive all the "subfolders" for the sections
		urls, err := DeriveCurseForgeURLs(projectURL)
		if err != nil {
			return nil, err
		}

		// Fetch & parse the sections subsequently, only parsing the header on the first call
		doHeader := true
		for section, url := range urls {
			// Only load specified sections
			if sections.Has(section) {
				// Fetch
				resp, err := FetchPage(url.String())
				if err != nil {
					return nil, fmt.Errorf("Error fetching URL '%s': %s", url.String(), err.Error())
				}
				// Parse
				err = results.ParseCurseForge(url, resp, doHeader, section, options)
				if err != nil {
					return nil, fmt.Errorf("Error parsing URL '%s': %s", url.String(), err.Error())
				}
				// Skip header on all subsequent calls
				doHeader = false
			}
		}
	}
	return results, nil
}

// ParseCurseForge parses single from curseforge.com
// (minecraft.curseforge.com, or feedthebeast.com, or or or - For a complete list, see curseforge.com).
//
// Supported & tested examples: see FetchCurseForge()
//
// results: The struct passed in results is filled with the parsed data.
// parseHeader: true, if the header values shall be parsed.
// section: A SINGLE section to tell which parser to use.
func (results *CurseForge) ParseCurseForge(documentURL *url.URL, resp *http.Response, parseHeader bool, section CurseForgeSections, options CurseForgeOptions) error {
	defer resp.Body.Close()

	root, err := xmlpath.ParseHTML(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing xml/http: %s", err.Error())
	}

	if parseHeader {
		err = parseCFHeader(results, documentURL, root, options)
		if err != nil {
			return fmt.Errorf("error processing CF header: %s", err.Error())
		}
	}

	switch section {
	case CFSectionOverview:
		err = parseCFOverview(results, documentURL, root, options)
		if err != nil {
			return fmt.Errorf("error processing CF Overview: %s", err.Error())
		}
	case CFSectionFiles:
		err = parseCFFiles(results, documentURL, root, options)
		if err != nil {
			return fmt.Errorf("error processing CF Files: %s", err.Error())
		}
	case CFSectionImages:
		err = parseCFImages(results, documentURL, root, options)
		if err != nil {
			return fmt.Errorf("error processing CF Files: %s", err.Error())
		}
	}

	return nil
}

func parseCFHeader(results *CurseForge, documentURLParsed *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
	var ok bool
	var err error

	var navbar *xmlpath.Node
	navbar, ok = pathCache.Node(root, "//nav[@class='e-header-nav']")
	if !ok {
		return fmt.Errorf("did not find navbar")
	}

	results.OverviewURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Overview')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Overview URL': %s", err.Error())
	}

	results.FilesURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Files')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Files URL': %s", err.Error())
	}

	results.ImagesURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Images')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Images URL': %s", err.Error())
	}

	results.IssuesURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Issues')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Issues URL': %s", err.Error())
	}

	// can be empty / non-present
	results.WikiURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Wiki')]/@href", documentURLParsed)
	/*if err != nil {
		return fmt.Errorf("error resolving value 'Wiki URL': %s", err.Error())
	}*/

	// can be empty / non-present
	results.SourceURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Source')]/@href", documentURLParsed)
	/*if err != nil {
		return fmt.Errorf("error resolving value 'Source URL': %s", err.Error())
	}*/

	results.DependenciesURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Dependencies')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Dependencies URL': %s", err.Error())
	}

	results.DependentsURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Dependents')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Dependents URL': %s", err.Error())
	}

	// Game (Actually: "Which curseforge is this?")
	results.Game, ok = pathCache.String(root, "//*[@id='site-main']/header//h1")
	if !ok {
		return fmt.Errorf("error resolving value 'Game'")
	}
	results.Game = strings.TrimSuffix(results.Game, " CurseForge")

	// Game URL
	results.GameURL, err = pathCache.URLWithBaseURL(root, "//*[@id='site-main']/header//a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Game URL': %s", err.Error())
	}

	var atf *xmlpath.Node
	atf, ok = pathCache.Node(root, "//*[@id='site-main']/section[@class='atf']")
	if !ok {
		return fmt.Errorf("did not find atf section")
	}

	// Title
	results.Title, ok = pathCache.String(atf, "//h1/a/span")
	if !ok {
		return fmt.Errorf("error resolving value 'Title'")
	}

	// Project URL
	results.ProjectURL, err = pathCache.URLWithBaseURL(atf, "//h1/a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Project URL': %s", err.Error())
	}

	// RootGameCategory
	results.RootGameCategory, ok = pathCache.String(atf, "//h2/a")
	if !ok {
		return fmt.Errorf("error resolving value 'RootGameCategory'")
	}

	// RootGameCategoryURL
	results.RootGameCategoryURL, err = pathCache.URLWithBaseURL(atf, "//h2/a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'RootGameCategoryURL': %s", err.Error())
	}

	// Avatar Image URL
	results.ImageURL, err = pathCache.URLWithBaseURL(atf, "//div[@class='avatar-wrapper']/a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'ImageURL': %s", err.Error())
	}
	// Avatar Image Thumbnail URL
	results.ImageThumbnailURL, err = pathCache.URLWithBaseURL(atf, "//div[@class='avatar-wrapper']/a/img/@src", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'ImageThumbnailURL': %s", err.Error())
	}
	// Donation URL
	// can be empty / non-present
	results.DontationURL, err = pathCache.URL(atf, "//a[@class='button tip icon-donate icon-paypal']/@href")
	/*if err != nil {
		return fmt.Errorf("error resolving value 'DontationURL': %s", err.Error())
	}*/

	return nil
}

func parseCFOverview(results *CurseForge, documentURL *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
	var ok bool
	var err error

	var sidebar *xmlpath.Node
	sidebar, ok = pathCache.Node(root, "//*[@id='content']/section/div[@class='e-project-details-secondary']")
	if !ok {
		return fmt.Errorf("did not find atf section: %s", err.Error())
	}

	/*
		Sidebar Values
	*/

	results.Created, err = pathCache.UnixTimestamp(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='Created ']/div[@class='info-data']/abbr/@data-epoch")
	if err != nil {
		return fmt.Errorf("error resolving value 'Created': %s", err.Error())
	}

	results.Updated, err = pathCache.UnixTimestamp(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='Last Released File ']/div[@class='info-data']/abbr/@data-epoch")
	if err != nil {
		return fmt.Errorf("error resolving value 'Updated // Last Released File': %s", err.Error())
	}

	results.TotalDownloads, err = pathCache.UInt(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='Total Downloads ']/div[@class='info-data']")
	if err != nil {
		return fmt.Errorf("error resolving value 'TotalDownloads': %s", err.Error())
	}

	results.License, ok = pathCache.String(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='License ']/div[@class='info-data']/a")
	if !ok {
		return fmt.Errorf("error resolving value 'License'")
	}

	results.LicenseURL, err = pathCache.URLWithBaseURL(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='License ']/div[@class='info-data']/a/@href", documentURL)
	if err != nil {
		return fmt.Errorf("error resolving value 'LicenseURL': %s", err.Error())
	}

	/*
		Categories
	*/

	categories := pathCache.Iter(sidebar, "//ul[@class='cf-details project-categories']/li")
	for categories.Next() {
		categoryNode := categories.Node()

		category := Category{}

		category.Name, ok = pathCache.String(categoryNode, "a/@title")
		if !ok {
			return fmt.Errorf("error resolving value 'Category/Name'")
		}

		category.URL, err = pathCache.URLWithBaseURL(categoryNode, "a/@href", documentURL)
		if err != nil {
			return fmt.Errorf("error resolving value 'Category/URL': %s", err.Error())
		}

		category.ImageURL, err = pathCache.URLWithBaseURL(categoryNode, "a/img/@src", documentURL)
		if err != nil {
			return fmt.Errorf("error resolving value 'Category/ImageURL': %s", err.Error())
		}

		results.Categories = append(results.Categories, category)

	}

	/*
		Links
	*/

	results.CurseURL, err = pathCache.URLWithBaseURL(sidebar, "//li[@class='view-on-curse']/a/@href", documentURL)
	if err != nil {
		return fmt.Errorf("error resolving value 'CurseURL': %s", err.Error())
	}
	//TODO: extract the curse project ID from the URL

	results.ReportProjectURL, err = pathCache.URLWithBaseURL(sidebar, "//li[@class='report-project']/a/@href", documentURL)
	if err != nil {
		return fmt.Errorf("error resolving value 'ReportProjectURL': %s", err.Error())
	}

	/*
		Members
	*/

	members := pathCache.Iter(sidebar, "//ul[@class='cf-details project-members']/li")
	for members.Next() {
		memberNode := members.Node()

		author := Author{}

		author.Name, ok = pathCache.String(memberNode, "div[@class='info-wrapper']/p/a[1]/span")
		if !ok {
			return fmt.Errorf("error resolving value 'Author/Name'")
		}

		author.URL, err = pathCache.URLWithBaseURL(memberNode, "div[@class='info-wrapper']/p/a[1]/@href", documentURL)
		if err != nil {
			return fmt.Errorf("error resolving value 'Author/URL': %s", err.Error())
		}

		author.Role, ok = pathCache.String(memberNode, "div[@class='info-wrapper']/p/span[@class='title']")
		if !ok {
			return fmt.Errorf("error resolving value 'Author/Role'")
		}

		author.ImageURL, err = pathCache.URLWithBaseURL(memberNode, "div/div/a/img/@src", documentURL)
		if err != nil {
			return fmt.Errorf("error resolving value 'Author/ImageURL': %s", err.Error())
		}

		results.Authors = append(results.Authors, author)

	}

	/*
		Recent Files
	*/
	if options.Has(CFOptionOverviewRecentFiles) {
		recents := pathCache.Iter(sidebar, "//div[@class='cf-sidebar-wrapper']//li[@class='file-tag']")
		for recents.Next() {
			fileTag := recents.Node()

			file := File{}

			file.ReleaseType, ok = pathCache.String(fileTag, "div[@class='e-project-file-phase-wrapper']/div/@title")
			if !ok {
				return fmt.Errorf("error resolving value 'File/ReleaseType'")
			}

			file.DirectURL, err = pathCache.URLWithBaseURL(fileTag, "//div[@class='project-file-download-button']/a/@href", documentURL)
			if err != nil {
				return fmt.Errorf("error resolving value 'File/DirectURL': %s", err.Error())
			}

			file.URL, err = pathCache.URLWithBaseURL(fileTag, "//div[@class='project-file-name-container']/a/@href", documentURL)
			if err != nil {
				return fmt.Errorf("error resolving value 'File/URL': %s", err.Error())
			}

			file.Name, ok = pathCache.String(fileTag, "//div[@class='project-file-name-container']/a/text()")
			if !ok {
				return fmt.Errorf("error resolving value 'File/Name'")
			}

			file.Date, err = pathCache.UnixTimestamp(fileTag, "//abbr/@data-epoch")
			if err != nil {
				return fmt.Errorf("error resolving value 'File/Date': %s", err.Error())
			}

			results.Downloads = append(results.Downloads, file)
		}
	}

	return nil
}

func parseCFFiles(results *CurseForge, documentURL *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {

	// Parse the files on the first page
	err := parseCFFilesSinglePage(results, documentURL, root, options)
	if err != nil {
		return fmt.Errorf("error parsing first files page: %s", err.Error())
	}

	// Stop if no pagination is requested
	if options.Has(CFOptionFilesNoPagination) {
		return nil
	}

	// Get the number of pages
	// (Last page is definitely listed as single element, so we just look for the one with the highest number)
	// (Could be optimized probably..)
	pagination := pathCache.Iter(root, "//div[@class='listing-header']//a[@class='b-pagination-item']")
	var pageCount uint64
	for pagination.Next() {
		pNode := pagination.Node()
		val, err := ParseUInt(pNode.String())
		if err != nil {
			return fmt.Errorf("error parsing page number: %s", err.Error())
		}
		// Just to be sure, compare if it is actually larger...
		if val > pageCount {
			pageCount = val
		}
	}

	// Sequentially, load the file pages
	var page uint64
	for page = 2; page <= pageCount; page++ {
		resp, err := FetchPage(documentURL.ResolveReference(&url.URL{
			Path:     "files",
			RawQuery: fmt.Sprintf("page=%d", page),
		}).String())
		if err != nil {
			return fmt.Errorf("error fetching subsequent files page (%d): %s", page, err.Error())
		}

		root, err := xmlpath.ParseHTML(resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing xml/http for subsequent files page (%d): %s", page, err.Error())
		}

		err = parseCFFilesSinglePage(results, documentURL, root, options)
		if err != nil {
			return fmt.Errorf("error parsing first files page: %s", err.Error())
		}
	}

	return nil
}

func parseCFFilesSinglePage(results *CurseForge, documentURL *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
	var ok bool
	var err error

	recents := pathCache.Iter(root, "//tr[@class='project-file-list-item']")
	for recents.Next() {
		fileTag := recents.Node()

		file := File{}

		file.ReleaseType, ok = pathCache.String(fileTag, "td[@class='project-file-release-type']/div/@title")
		if !ok {
			return fmt.Errorf("error resolving value 'File/ReleaseType'")
		}

		file.DirectURL, err = pathCache.URLWithBaseURL(fileTag, "//div[@class='project-file-download-button']/a/@href", documentURL)
		if err != nil {
			return fmt.Errorf("error resolving value 'File/DirectURL': %s", err.Error())
		}

		file.URL, err = pathCache.URLWithBaseURL(fileTag, "//div[@class='project-file-name-container']/a/@href", documentURL)
		if err != nil {
			return fmt.Errorf("error resolving value 'File/URL': %s", err.Error())
		}

		file.Name, ok = pathCache.String(fileTag, "//div[@class='project-file-name-container']/a/text()")
		if !ok {
			return fmt.Errorf("error resolving value 'File/Name'")
		}

		_, ok = pathCache.String(fileTag, "//div[@class='project-file-name-container']/a[@class='more-files-tag']")
		file.HasAdditionalFiles = ok

		file.SizeInfo, ok = pathCache.String(fileTag, "//td[@class='project-file-size']/text()")
		if !ok {
			return fmt.Errorf("error resolving value 'File/SizeInfo'")
		}

		file.Date, err = pathCache.UnixTimestamp(fileTag, "//abbr/@data-epoch")
		if err != nil {
			return fmt.Errorf("error resolving value 'File/Date': %s", err.Error())
		}

		file.GameVersion, ok = pathCache.String(fileTag, "//span[@class='version-label']/text()")
		if !ok {
			return fmt.Errorf("error resolving value 'File/GameVersion'")
		}

		file.Downloads, err = pathCache.UInt(fileTag, "//td[@class='project-file-downloads']/text()")
		if err != nil {
			return fmt.Errorf("error resolving value 'File/Downloads': %s", err.Error())
		}

		results.Downloads = append(results.Downloads, file)
	}
	return nil
}

func parseCFImages(results *CurseForge, documentURL *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
	//var ok bool
	//var err error

	//TODO: implement

	return nil
}
