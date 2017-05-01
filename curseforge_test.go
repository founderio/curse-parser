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
	"strings"
	"testing"
	"time"
)

func TestDeriveCurseForgeURLs(t *testing.T) {
	projectURL, err := url.Parse("https://minecraft.curseforge.com/projects/taam")
	if err != nil {
		t.Fatal("Parsing rest URL", err.Error())
	}
	urls, err := DeriveCurseForgeURLs(projectURL)
	if err != nil {
		t.Fatal(err.Error())
	}
	if "https://minecraft.curseforge.com/projects/taam/files" != urls[CFSectionFiles].String() {
		t.Errorf("Expected '%s', got '%s'", "https://minecraft.curseforge.com/projects/taam/files", urls[CFSectionFiles].String())
	}
	if "https://minecraft.curseforge.com/projects/taam/images" != urls[CFSectionImages].String() {
		t.Errorf("Expected '%s', got '%s'", "https://minecraft.curseforge.com/projects/taam/images", urls[CFSectionFiles].String())
	}
}

func TestParseCurseForge(t *testing.T) {
	testUrls := []string{
		"https://minecraft.curseforge.com/projects/taam",
		"https://wow.curseforge.com/projects/pawn",
	}

	for idx, tURL := range testUrls {
		resp, err := FetchPage(tURL)
		if err != nil {
			t.Fatal(err)
		}
		var pURL *url.URL
		pURL, err = url.Parse(tURL)
		if err != nil {
			t.Fatal(err)
		}

		results := new(CurseForge)

		err = results.ParseCurseForge(pURL, resp, true, CFSectionOverview, CFOptionNone)
		if err != nil {
			t.Fatal(err)
		}

		// For first element (taam project page), check for existence of the donation URL.
		validateResultsCurseforge(t, tURL, results, idx == 0)

	}
}

func TestFetchCurseForge(t *testing.T) {
	testUrls := []string{
		"https://minecraft.curseforge.com/projects/taam",
		"https://wow.curseforge.com/projects/pawn",
	}

	for idx, tURL := range testUrls {
		var pURL *url.URL
		pURL, err := url.Parse(tURL)
		if err != nil {
			t.Fatal(err)
		}

		results, err := FetchCurseForge(pURL, CFSectionOverview|CFSectionFiles|CFSectionImages, CFOptionNone)
		if err != nil {
			t.Fatal(err)
		}

		// For first element (taam project page), check for existence of the donation URL.
		validateResultsCurseforge(t, tURL, results, idx == 0)

	}
}

func validateResultsCurseforge(t *testing.T, url string, results *CurseForge, expectDonationURL bool) {

	if results.OverviewURL == nil || results.OverviewURL.Host == "" {
		t.Errorf("Empty value 'OverviewURL' when testing URL %s", url)
	}
	if results.FilesURL == nil || results.FilesURL.Host == "" {
		t.Errorf("Empty value 'FilesURL' when testing URL %s", url)
	}
	if results.ImagesURL == nil || results.ImagesURL.Host == "" {
		t.Errorf("Empty value 'ImagesURL' when testing URL %s", url)
	}
	if results.DependenciesURL == nil || results.DependenciesURL.Host == "" {
		t.Errorf("Empty value 'DependenciesURL' when testing URL %s", url)
	}
	if results.DependentsURL == nil || results.DependentsURL.Host == "" {
		t.Errorf("Empty value 'DependentsURL' when testing URL %s", url)
	}

	if results.CurseURL == nil || results.CurseURL.Host == "" {
		t.Errorf("Empty value 'CurseURL' when testing URL %s", url)
	}
	if results.ReportProjectURL == nil || results.ReportProjectURL.Host == "" {
		t.Errorf("Empty value 'ReportProjectURL' when testing URL %s", url)
	}
	//TODO: optionally test these (based on expected values, like donation URL)
	/*if results.IssuesURL == nil || results.IssuesURL.Host == "" {
		t.Errorf("Empty value 'IssuesURL' when testing URL %s", url)
	}
	if results.WikiURL == nil || results.WikiURL.Host == "" {
		t.Errorf("Empty value 'WikiURL' when testing URL %s", url)
	}
	if results.SourceURL == nil || results.SourceURL.Host == "" {
		t.Errorf("Empty value 'SourceURL' when testing URL %s", url)
	}*/

	if results.Title == "" {
		t.Errorf("Empty value 'Title' when testing URL %s", url)
	}
	if results.ProjectURL == nil || results.ProjectURL.Host == "" {
		t.Errorf("Empty value 'GameURL' when testing URL %s", url)
	}
	if expectDonationURL {
		if results.DontationURL == nil || results.DontationURL.Host == "" {
			t.Errorf("Empty value 'DontationURL' when testing URL %s", url)
		}
	}
	if results.ImageURL == nil || results.ImageURL.Host == "" {
		t.Errorf("Empty value 'RootGameCategoryURL' when testing URL %s", url)
	}
	if results.ImageThumbnailURL == nil || results.ImageThumbnailURL.Host == "" {
		t.Errorf("Empty value 'RootGameCategoryURL' when testing URL %s", url)
	}

	if results.GameURL == nil || results.GameURL.Host == "" {
		t.Errorf("Empty value 'GameURL' when testing URL %s", url)
	}
	if results.RootGameCategory == "" {
		t.Errorf("Empty value 'RootGameCategory' when testing URL %s", url)
	}
	if results.RootGameCategoryURL == nil || results.RootGameCategoryURL.Host == "" {
		t.Errorf("Empty value 'RootGameCategoryURL' when testing URL %s", url)
	}
	if results.License == "" {
		t.Errorf("Empty value 'License' when testing URL %s", url)
	}
	if results.LicenseURL == nil || results.LicenseURL.Host == "" {
		t.Errorf("Empty value 'LicenseURL' when testing URL %s", url)
	}
	if results.Game == "" {
		t.Errorf("Empty value 'Game' when testing URL %s", url)
	}
	if results.GameURL == nil || results.GameURL.Host == "" {
		t.Errorf("Empty value 'GameURL' when testing URL %s", url)
	}
	if results.TotalDownloads == 0 {
		t.Errorf("Empty value 'TotalDownloads' when testing URL %s", url)
	}

	if results.Created == time.Unix(0, 0).UTC() {
		t.Errorf("Empty value 'Created' when testing URL %s", url)
	}
	if results.Updated == time.Unix(0, 0).UTC() {
		t.Errorf("Empty value 'Updated' when testing URL %s", url)
	}

	if len(results.Authors) == 0 {
		t.Errorf("Empty list 'Authors' when testing URL %s", url)
	}
	for _, a := range results.Authors {
		if a.Name == "" {
			t.Errorf("Empty value 'Author/Name' when testing URL %s", url)
		}
		if a.Role == "" {
			t.Errorf("Empty value 'Author/Role' when testing URL %s", url)
		}
		if strings.Contains(a.Role, ":") {
			t.Errorf("Trimming ':' from author role failed when testing URL %s", url)
		}
		if a.Role == "" {
			t.Errorf("Empty value 'Author/Role' when testing URL %s", url)
		}
		if a.URL == nil || a.URL.Host == "" {
			t.Errorf("Empty value 'Author/URL' when testing URL %s", url)
		}
		if a.ImageURL == nil || a.ImageURL.Host == "" {
			t.Errorf("Empty value 'Author/ImageURL' when testing URL %s", url)
		}
	}

	if len(results.Categories) == 0 {
		t.Errorf("Empty list 'Categories' when testing URL %s", url)
	}
	for _, c := range results.Categories {

		if c.URL == nil || c.URL.Host == "" {
			t.Errorf("Empty value 'Category/URL' when testing URL %s", url)
		}

		if c.ImageURL == nil || c.ImageURL.Host == "" {
			t.Errorf("Empty value 'Category/ImageURL' when testing URL %s", url)
		}

		if c.Name == "" {
			t.Errorf("Empty value 'Category/Name' when testing URL %s", url)
		}
	}
}
