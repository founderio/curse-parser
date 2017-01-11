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

	"gopkg.in/xmlpath.v2"
)

// Fetch & parse mod pages from curseforge.com
// (minecraft.curseforge.com, or feed-the-beast.com, or or or - For a complete list, see curseforge.com).
//
// This function expects to get the project URL (e.g. "https://minecraft.curseforge.com/projects/taam") and will build
// all other required URLs based on the content selection.
//
// Other parameters:
//
// sections: Defines the content pages to be fetched & parsed.
//      If 0 (CFHeader) is passed only, the overview page
//      is loaded and only the header values are parsed & returned.
//      Multiple values can be added to select multiple sections, e.g. CFSectionFiles | CFSectionImages
//      This causes multiple sequential or parallel requests to CurseForge.
// options: Pass CFOptionNone, or one or more of the other options to tweak some behaviour of the content parsers.
//      Multiple values can be added to define multiple options, e.g. CFOptionOverviewRecentFiles | CFOptionFilesNoPagination
// parallel: Select whether to make parallel or sequential calls.
//      Pass true to load all selected pages simultaneously.
//      Disclaimer: Be aware that CurseForge might impose rate limiting on you should you overdo the parallelism.
func FetchCurseForge(projectURL *url.URL, sections CurseForgeSections, options CurseForgeOptions, parallel bool) (*CurseforgeDotCom, error) {
	var err error
	results := new(CurseforgeDotCom)

	if sections == CFSectionHeader {
		//TODO: WHY does that require a string??
		var resp *http.Response

		resp, err = FetchPage(projectURL.String())
		if err != nil {
			return nil, err
		}
		err = ParseCurseForge(projectURL, resp, results, true, CFSectionHeader, options)

	} else {
		//TODO: Build required URLs & fetch
	}
	return results, nil
}

// Parse mod pages from curseforge.com
// (minecraft.curseforge.com, or feedthebeast.com, or or or - For a complete list, see curseforge.com).
//
// Supported & tested examples:
// * https://minecraft.curseforge.com/projects/taam
//
// results: The struct passed in results is filled with the parsed data.
// parseHeader: true, if the header values shall be parsed.
// section: A SINGLE section to tell which parser to use.
func ParseCurseForge(documentURL *url.URL, resp *http.Response, results *CurseforgeDotCom, parseHeader bool, section CurseForgeSections, options CurseForgeOptions) error {
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
	}

	//TODO: determine where we are and parse the correct part

	return nil
}

func parseCFHeader(results *CurseforgeDotCom, documentURLParsed *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
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

	results.WikiURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Wiki')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Wiki URL': %s", err.Error())
	}

	results.SourceURL, err = pathCache.URLWithBaseURL(navbar, "//li/a[contains(text(), 'Source')]/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'Source URL': %s", err.Error())
	}

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
	results.DontationURL, err = pathCache.URL(atf, "//a[@class='button tip icon-donate icon-paypal']/@href")
	if err != nil {
		return fmt.Errorf("error resolving value 'DontationURL': %s", err.Error())
	}

	return nil
}

func parseCFOverview(results *CurseforgeDotCom, documentURL *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
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

func parseCFFiles(results *CurseforgeDotCom, documentURL *url.URL, root *xmlpath.Node, options CurseForgeOptions) error {
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


		file.SizeInfo, ok = pathCache.String(fileTag, "//td[@class='project-file-size']/a/text()")
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

	//TODO: pagination

	return nil
}