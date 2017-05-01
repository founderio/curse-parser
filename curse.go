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

	"gopkg.in/xmlpath.v2"
)

// ParseCurse parses mod pages from mods.curse.com.
// Supported & tested examples:
// * https://mods.curse.com/mc-mods/minecraft/238424-taam
// * https://mods.curse.com/texture-packs/minecraft/equanimity-32x
// * https://mods.curse.com/worlds/minecraft/246026-skyblock-3
// * https://mods.curse.com/addons/wow/pawn
func ParseCurse(documentURL string, resp *http.Response) (*Curse, error) {
	defer resp.Body.Close()

	documentURLParsed, err := url.Parse(strings.TrimSpace(documentURL))
	if err != nil {
		return nil, err
	}

	root, err := xmlpath.ParseHTML(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing xml/http: %s", err.Error())
	}

	results := new(Curse)

	var ok bool
	// Temp-Variable for values to be parsed
	var parseString string

	/*
		Get the project-overview node for faster processing
	*/
	projectOverview, ok := pathCache.Node(root, "//*[@id='project-overview']")
	if !ok {
		return nil, errors.New("could not find 'project-overview' in response body")
	}

	// Title
	results.Title, ok = pathCache.String(projectOverview, "header/h2")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Title'")
	}

	// Donation Link
	results.DontationURL, err = pathCache.URL(projectOverview, "div[@class='meta-info']/div/a/@href")
	if err != nil {
		// Some projects do not have a donation URL -> don't fail!
		results.DontationURL = nil
	}

	// Authors

	iter := pathCache.Iter(projectOverview, "div[@class='main-details']/div[@class='main-info']/ul[@class='authors group']/li")
	for iter.Next() {
		authorNode := iter.Node()
		author := Author{}

		// Author Name
		author.Name, ok = pathCache.String(authorNode, "a")
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Author/Name'")
		}

		// Author Role
		author.Role, ok = pathCache.String(authorNode, "text()")
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Author/Role'")
		}
		author.Role = strings.TrimSuffix(author.Role, ":")

		// Link to author's page
		author.URL, err = pathCache.URLWithBaseURL(authorNode, "a/@href", documentURLParsed)
		if err != nil {
			return nil, fmt.Errorf("error resolving value 'Author/URL': %s", err.Error())
		}

		results.Authors = append(results.Authors, author)
	}

	// Categories

	iter = pathCache.Iter(projectOverview, "div[@class='main-details']/div[@class='main-info']/a")
	for iter.Next() {
		categoryNode := iter.Node()
		category := Category{}

		// Author Name
		category.Name, ok = pathCache.String(categoryNode, "@title")
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Category/Name'")
		}

		// Link to category page
		category.URL, err = pathCache.URLWithBaseURL(categoryNode, "@href", documentURLParsed)
		if err != nil {
			return nil, fmt.Errorf("error resolving value 'Category/URL': %s", err.Error())
		}

		// Link to category image
		category.ImageURL, err = pathCache.URL(categoryNode, "img/@src")
		if err != nil {
			return nil, fmt.Errorf("error resolving value 'Category/ImageURL': %s", err.Error())
		}

		results.Categories = append(results.Categories, category)
	}

	// Likes
	parseString, ok = pathCache.String(projectOverview, "div[@class='main-details']/div[@class='main-info']/div[@class='appreciate']/ul/li[@class='grats']/span")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Likes'")
	}
	// Format of this value: "nnn Likes" -> get the first 'field'
	parseString = strings.Fields(parseString)[0]
	results.Likes, err = strconv.ParseUint(parseString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Likes': %s", err.Error())
	}

	/*
		Get the details-list node for faster processing
	*/
	detailsList, ok := pathCache.Node(projectOverview, "div/div/ul[@class='details-list']")
	if !ok {
		return nil, errors.New("could not find 'details-list' in response body")
	}

	// Game
	results.Game, ok = pathCache.String(detailsList, "li[@class='game']")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Game'")
	}

	// URL
	results.GameURL, err = pathCache.URLWithBaseURL(detailsList, "li[@class='game']/a/@href", documentURLParsed)
	if err != nil {
		return nil, fmt.Errorf("error resolving value 'GameURL': %s", err.Error())
	}

	// Average Downloads
	parseString, ok = pathCache.String(detailsList, "li[@class='average-downloads']")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Average Downloads'")
	}
	// Format of this value: "nnn Monthly Downloads" -> get the first & second 'field'
	split := strings.Fields(parseString)
	results.AvgDownloads, err = strconv.ParseUint(strings.Replace(split[0], ",", "", -1), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Average Downloads': %s", err.Error())
	}
	results.AvgDownloadsTimeframe = split[1]

	// Total Downloads
	parseString, ok = pathCache.String(detailsList, "li[@class='downloads']")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Total Downloads'")
	}
	// Format of this value: "nnn Total Downloads" -> get the first 'field'
	parseString = strings.Fields(parseString)[0]
	results.TotalDownloads, err = strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Total Downloads': %s", err.Error())
	}

	// Updated
	results.Updated, err = pathCache.UnixTimestamp(detailsList, "li[@class='updated' and text()='Updated ']/abbr/@data-epoch")
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Updated': %s", err.Error())
	}

	// Created
	results.Created, err = pathCache.UnixTimestamp(detailsList, "li[@class='updated' and text()='Created ']/abbr/@data-epoch")
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Created': %s", err.Error())
	}

	// Favorites
	parseString, ok = pathCache.String(detailsList, "li[@class='favorited']")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Favorites'")
	}
	// Format of this value: "nnn Favorites" -> get the first 'field'
	parseString = strings.Fields(parseString)[0]
	results.Favorites, err = strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Favorites': %s", err.Error())
	}

	// Project Site // Curseforge URL
	results.CurseforgeURL, err = pathCache.URL(detailsList, "li[@class='curseforge']/a/@href")
	if err != nil {
		return nil, fmt.Errorf("error parsing URL for 'Curseforge URL': %s", err.Error())
	}

	// License
	parseString, ok = pathCache.String(detailsList, "li[@class='license']")
	if !ok {
		return nil, fmt.Errorf("error resolving value 'License'")
	}
	results.License = strings.TrimPrefix(parseString, "License: ")

	// Screenshots
	iter = pathCache.Iter(root, "//div[@id='screenshot-gallery']//div[@class='listing-body']/ul/li/a")
	for iter.Next() {
		screenshotNode := iter.Node()
		screenshot := Image{}

		// URL
		screenshot.URL, err = pathCache.URL(screenshotNode, "@href")
		if err != nil {
			return nil, fmt.Errorf("error parsing URL for 'Screenshot/URL': %s", err.Error())
		}

		results.Screenshots = append(results.Screenshots, screenshot)
	}

	// Downloads
	iter = pathCache.Iter(root, "//div[@id='tab-other-downloads']//div[@class='listing-body']/table/tbody/tr")
	for iter.Next() {
		downloadNode := iter.Node()
		download := File{}

		// Name
		download.Name, ok = pathCache.String(downloadNode, "td[1]/a")
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/Name'")
		}

		// URL
		download.URL, err = pathCache.URLWithBaseURL(downloadNode, "td[1]/a/@href", documentURLParsed)
		if err != nil {
			return nil, fmt.Errorf("error resolving value 'Download/URL': %s", err.Error())
		}

		// ReleaseType
		download.ReleaseType, ok = pathCache.String(downloadNode, "td[2]")
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/ReleaseType'")
		}

		// GameVersion
		download.GameVersion, ok = pathCache.String(downloadNode, "td[3]")
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/GameVersion'")
		}

		// Downloads
		download.Downloads, err = pathCache.UInt(downloadNode, "td[4]")
		if err != nil {
			return nil, fmt.Errorf("error parsing value for 'Download/Downloads': %s", err.Error())
		}

		// Date
		download.Date, err = pathCache.UnixTimestamp(downloadNode, "td[5]/abbr/@data-epoch")
		if err != nil {
			return nil, fmt.Errorf("error parsing number for 'Download/Date': %s", err.Error())
		}

		results.Downloads = append(results.Downloads, download)
	}

	return results, nil
}
