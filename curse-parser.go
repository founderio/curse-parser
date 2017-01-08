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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gopkg.in/xmlpath.v2"
)

func ParseModsDotCurseDotCom(documentURL string, resp *http.Response) (*ModsDotCurseDotCom, error) {
	defer resp.Body.Close()

	root, err := xmlpath.ParseHTML(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing xml/http: %s", err.Error())
	}

	results := new(ModsDotCurseDotCom)

	var ok bool
	// Temp-Variable for values to be parsed
	var parseString string

	//TODO: extract the xpath compilation to a central place

	/*
		Get the project-overview node for faster processing
	*/
	pathProjectOverview := xmlpath.MustCompile("//*[@id='project-overview']")
	iter := pathProjectOverview.Iter(root)
	if !iter.Next() {
		return nil, errors.New("could not find 'project-overview' in response body")
	}
	projectOverview := iter.Node()

	// Title
	title := xmlpath.MustCompile("header/h2")
	results.Title, ok = title.String(projectOverview)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Title'")
	}

	// Donation Link
	donation := xmlpath.MustCompile("div[@class='meta-info']/div/a/@href")
	results.DontationLink, ok = donation.String(projectOverview)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'DonationLink'")
	}

	// Authors
	authors := xmlpath.MustCompile("div[@class='main-details']/div[@class='main-info']/ul[@class='authors group']/li")
	role := xmlpath.MustCompile("text()")
	authorName := xmlpath.MustCompile("a")
	relativeLink := xmlpath.MustCompile("a/@href")

	iter = authors.Iter(projectOverview)
	for iter.Next() {
		authorNode := iter.Node()
		author := Author{}

		// Author Name
		author.Name, ok = authorName.String(authorNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Author/Name'")
		}

		// Author Role
		author.Role, ok = role.String(authorNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Author/Role'")
		}
		author.Role = strings.TrimSuffix(author.Role, ": ")

		// Link to author's page
		parseString, ok = relativeLink.String(authorNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Author/URL'")
		}
		author.URL, err = combineUrl(documentURL, parseString)
		if err != nil {
			return nil, fmt.Errorf("error parsing value 'Author/URL': %s", err.Error())
		}
		results.Authors = append(results.Authors, author)
	}

	// Categories
	categories := xmlpath.MustCompile("div[@class='main-details']/div[@class='main-info']/a")
	categoryName := xmlpath.MustCompile("@title")
	categoryURL := xmlpath.MustCompile("@href")
	categoryImageURL := xmlpath.MustCompile("img/@src")

	iter = categories.Iter(projectOverview)
	for iter.Next() {
		categoryNode := iter.Node()
		category := Category{}

		// Author Name
		category.Name, ok = categoryName.String(categoryNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Category/Name'")
		}

		// Link to category page
		parseString, ok = categoryURL.String(categoryNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Category/URL'")
		}
		category.URL, err = combineUrl(documentURL, parseString)
		if err != nil {
			return nil, fmt.Errorf("error parsing value 'Author/Url': %s", err.Error())
		}

		// Link to category image
		parseString, ok = categoryImageURL.String(categoryNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Category/ImageURL'")
		}
		category.ImageURL, err = url.Parse(parseString)
		if err != nil {
			return nil, fmt.Errorf("error parsing value 'Category/ImageURL': %s", err.Error())
		}
		results.Categories = append(results.Categories, category)
	}

	// Likes
	likes := xmlpath.MustCompile("div[@class='main-details']/div[@class='main-info']/div[@class='appreciate']/ul/li[@class='grats']/span")
	parseString, ok = likes.String(projectOverview)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Likes'")
	}
	// Format of this value: "nnn Likes" -> get the first 'field'
	parseString = strings.Fields(strings.TrimSpace(parseString))[0]
	results.Likes, err = strconv.ParseUint(parseString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Likes': %s", err.Error())
	}

	/*
		Get the details-list node for faster processing
	*/
	pathDetailsList := xmlpath.MustCompile("div/div/ul[@class='details-list']")
	iter = pathDetailsList.Iter(projectOverview)
	if !iter.Next() {
		return nil, errors.New("could not find 'details-list' in response body")
	}
	detailsList := iter.Node()

	// Game
	game := xmlpath.MustCompile("li[@class='game']")
	results.Game, ok = game.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Game'")
	}

	// Average Downloads
	averageDownloads := xmlpath.MustCompile("li[@class='average-downloads']")
	parseString, ok = averageDownloads.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Average Downloads'")
	}
	// Format of this value: "nnn Monthly Downloads" -> get the first & second 'field'
	split := strings.Fields(strings.TrimSpace(parseString))
	results.AvgDownloads, err = strconv.ParseUint(strings.Replace(split[0], ",", "", -1), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Average Downloads': %s", err.Error())
	}
	results.AvgDownloadsTimeframe = split[1]

	// Total Downloads
	totalDownloads := xmlpath.MustCompile("li[@class='downloads']")
	parseString, ok = totalDownloads.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Total Downloads'")
	}
	// Format of this value: "nnn Total Downloads" -> get the first 'field'
	parseString = strings.Fields(strings.TrimSpace(parseString))[0]
	results.TotalDownloads, err = strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Total Downloads': %s", err.Error())
	}

	// Updated
	updated := xmlpath.MustCompile("li[@class='updated' and text()='Updated ']/abbr/@data-epoch")
	parseString, ok = updated.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Updated'")
	}
	// Format of this value: "nnn Total Downloads" -> get the first 'field'
	parseString = strings.Fields(strings.TrimSpace(parseString))[0]
	var ts int64
	ts, err = strconv.ParseInt(strings.TrimSpace(parseString), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Updated': %s", err.Error())
	}
	results.Updated = time.Unix(ts, 0).UTC()

	// Created
	created := xmlpath.MustCompile("li[@class='updated' and text()='Created ']/abbr/@data-epoch")
	parseString, ok = created.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Created'")
	}
	// Format of this value: "nnn Total Downloads" -> get the first 'field'
	parseString = strings.Fields(strings.TrimSpace(parseString))[0]
	ts, err = strconv.ParseInt(strings.TrimSpace(parseString), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Created': %s", err.Error())
	}
	results.Created = time.Unix(ts, 0).UTC()

	// Favorites
	favorites := xmlpath.MustCompile("li[@class='favorited']")
	parseString, ok = favorites.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Favorites'")
	}
	// Format of this value: "nnn Favorites" -> get the first 'field'
	parseString = strings.Fields(strings.TrimSpace(parseString))[0]
	results.Favorites, err = strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing number for 'Favorites': %s", err.Error())
	}

	// Project Site // Curseforge URL
	curseforge := xmlpath.MustCompile("li[@class='curseforge']/a/@href")
	parseString, ok = curseforge.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'Curseforge URL'")
	}
	results.CurseforgeURL, err = url.Parse(strings.TrimSpace(parseString))
	if err != nil {
		return nil, fmt.Errorf("error parsing URL for 'Curseforge URL': %s", err.Error())
	}
	// URL may be specified in schemeless format "//www.curseforge.com/..."
	if results.CurseforgeURL.Scheme == "" {
		results.CurseforgeURL.Scheme = "https"
	}

	// License
	license := xmlpath.MustCompile("li[@class='license']")
	parseString, ok = license.String(detailsList)
	if !ok {
		return nil, fmt.Errorf("error resolving value 'License'")
	}
	results.License = strings.TrimPrefix(strings.TrimSpace(parseString), "License: ")

	// Screenshots
	screenshots := xmlpath.MustCompile("//div[@id='screenshot-gallery']//div[@class='listing-body']/ul/li/a")
	srcImage := xmlpath.MustCompile("@href")

	iter = screenshots.Iter(root)
	for iter.Next() {
		screenshotNode := iter.Node()
		screenshot := Image{}

		// URL
		parseString, ok = srcImage.String(screenshotNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Screenshot/URL'")
		}
		screenshot.URL, err = url.Parse(strings.TrimSpace(parseString))
		if err != nil {
			return nil, fmt.Errorf("error parsing URL for 'Screenshot/URL': %s", err.Error())
		}

		results.Screenshots = append(results.Screenshots, screenshot)
	}

	// Downloads
	downloads := xmlpath.MustCompile("//div[@id='tab-other-downloads']//div[@class='listing-body']/table/tbody/tr")
	downloadName := xmlpath.MustCompile("td[1]/a")
	downloadURL := xmlpath.MustCompile("td[1]/a/@href")
	downloadReleaseType := xmlpath.MustCompile("td[2]")
	downloadGameVersion := xmlpath.MustCompile("td[3]")
	downloadDownloads := xmlpath.MustCompile("td[4]")
	downloadDate := xmlpath.MustCompile("td[5]/abbr/@data-epoch")

	iter = downloads.Iter(root)
	for iter.Next() {
		downloadNode := iter.Node()
		download := Download{}

		// Name
		download.Name, ok = downloadName.String(downloadNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/Name'")
		}

		// URL
		parseString, ok = downloadURL.String(downloadNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/URL'")
		}
		download.URL, err = combineUrl(documentURL, parseString)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL for 'Download/URL': %s", err.Error())
		}

		// ReleaseType
		download.ReleaseType, ok = downloadReleaseType.String(downloadNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/ReleaseType'")
		}

		// GameVersion
		download.GameVersion, ok = downloadGameVersion.String(downloadNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/GameVersion'")
		}

		// Downloads
		parseString, ok = downloadDownloads.String(downloadNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/Downloads'")
		}
		download.Downloads, err = strconv.ParseUint(strings.Replace(parseString, ",", "", -1), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing value for 'Download/Downloads': %s", err.Error())
		}

		// Date
		parseString, ok = downloadDate.String(downloadNode)
		if !ok {
			return nil, fmt.Errorf("error resolving value 'Download/Date'")
		}
		ts, err = strconv.ParseInt(strings.TrimSpace(parseString), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing number for 'Download/Date': %s", err.Error())
		}
		download.Date = time.Unix(ts, 0).UTC()

		results.Downloads = append(results.Downloads, download)
	}

	return results, nil
}
