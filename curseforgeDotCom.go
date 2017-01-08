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

// Parse mod pages from curseforge.com (or minecraft.curseforge.com).
// Supported & tested examples:
// * https://minecraft.curseforge.com/projects/taam
func ParseCurseforgeDotCom(documentURL string, resp *http.Response, parseHeader bool) (*CurseforgeDotCom, error) {
	defer resp.Body.Close()

	documentURLParsed, err := url.Parse(strings.TrimSpace(documentURL))
	if err != nil {
		return nil, err
	}

	root, err := xmlpath.ParseHTML(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing xml/http: %s", err.Error())
	}

	results := new(CurseforgeDotCom)

	if parseHeader {
		err = parseCFHeader(results, documentURLParsed, root)
		if err != nil {
			return nil, fmt.Errorf("error processing CF header: %s", err.Error())
		}
	}

	//TODO: determine where we are and parse the correct part

	err = parseCFOverview(results, documentURLParsed, root)
	if err != nil {
		return nil, fmt.Errorf("error processing CF Overview: %s", err.Error())
	}

	return results, nil
}

func parseCFHeader(results *CurseforgeDotCom, documentURLParsed *url.URL, root *xmlpath.Node) error {
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

func parseCFOverview(results *CurseforgeDotCom, documentURLParsed *url.URL, root *xmlpath.Node) error {
	var ok bool
	var err error

	var sidebar *xmlpath.Node
	sidebar, ok = pathCache.Node(root, "//*[@id='content']/section/div[@class='e-project-details-secondary']")
	if !ok {
		return fmt.Errorf("did not find atf section: %s", err.Error())
	}

	// Project ID is NOT part of the downloaded file for some reason...
	/*results.ProjectID, ok = pathCache.String(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='Project ID ']/div[@class='info-data']")
	if !ok {
		return fmt.Errorf("error resolving value 'ProjectID'")
	}*/

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
	results.License = strings.TrimSpace(results.License)

	results.LicenseURL, err = pathCache.URLWithBaseURL(sidebar, "//ul[@class='cf-details project-details']/li[div[@class='info-label']='License ']/div[@class='info-data']/a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'LicenseURL': %s", err.Error())
	}



	results.CurseURL, err = pathCache.URLWithBaseURL(sidebar, "//li[@class='view-on-curse']/a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'CurseURL': %s", err.Error())
	}

	results.ReportProjectURL, err = pathCache.URLWithBaseURL(sidebar, "//li[@class='report-project']/a/@href", documentURLParsed)
	if err != nil {
		return fmt.Errorf("error resolving value 'ReportProjectURL': %s", err.Error())
	}

	members := pathCache.Iter(sidebar, "//ul[@class='cf-details project-members']/li")
	for members.Next() {
		memberNode := members.Node()

		author := Author{}

		author.Name, ok = pathCache.String(memberNode, "div[@class='info-wrapper']/p/a[1]/span")
		if !ok {
			return fmt.Errorf("error resolving value 'Author/Name'")
		}

		author.URL, err = pathCache.URLWithBaseURL(memberNode, "div[@class='info-wrapper']/p/a[1]/@href", documentURLParsed)
		if err != nil {
			return fmt.Errorf("error resolving value 'Author/URL': %s", err.Error())
		}

		author.Role, ok = pathCache.String(memberNode, "div[@class='info-wrapper']/p/span[@class='title']")
		if !ok {
			return fmt.Errorf("error resolving value 'Author/Role'")
		}

		author.ImageURL, err = pathCache.URLWithBaseURL(memberNode, "div/div/a/img/@src", documentURLParsed)
		if err != nil {
			return fmt.Errorf("error resolving value 'Author/ImageURL': %s", err.Error())
		}

		results.Authors = append(results.Authors, author)

	}

	categories := pathCache.Iter(sidebar, "//ul[@class='cf-details project-categories']/li")
	for categories.Next() {
		categoryNode := categories.Node()

		category := Category{}

		category.Name, ok = pathCache.String(categoryNode, "a/@title")
		if !ok {
			return fmt.Errorf("error resolving value 'Category/Name'")
		}

		category.URL, err = pathCache.URLWithBaseURL(categoryNode, "a/@href", documentURLParsed)
		if err != nil {
			return fmt.Errorf("error resolving value 'Category/URL': %s", err.Error())
		}

		category.ImageURL, err = pathCache.URLWithBaseURL(categoryNode, "a/img/@src", documentURLParsed)
		if err != nil {
			return fmt.Errorf("error resolving value 'Category/ImageURL': %s", err.Error())
		}

		results.Categories = append(results.Categories, category)

	}

	return nil
}