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
	"testing"
	"time"
	"strings"
)

func TestParseModsDotCurseDotCom(t *testing.T) {
	testUrls := []string{
		"https://mods.curse.com/mc-mods/minecraft/238424-taam",
		"https://mods.curse.com/texture-packs/minecraft/equanimity-32x",
		"https://mods.curse.com/worlds/minecraft/246026-skyblock-3",
		"https://mods.curse.com/addons/wow/pawn",
	}

	for idx, url := range testUrls {
		resp, err := FetchPage(url)
		if err != nil {
			t.Fatal(err)
		}

		results, err := ParseModsDotCurseDotCom(url, resp)
		if err != nil {
			t.Fatal(err)
		}

		// For first element (taam project page), check for existence of the donation URL.
		validateResults(t, url, results, idx == 0)
	}
}

func validateResults(t *testing.T, url string, results *ModsDotCurseDotCom, expectDonationURL bool) {
	// Just some basic tests that tell us when a value returns nil or default values.
	// If that is the case, the parser is likely borked because curse changed their website layout.

	if len(results.Downloads) == 0 {
		t.Errorf("Empty list 'Downloads' when testing URL %s", url)
	} else {
		for _,dl := range results.Downloads {

			if dl.Date == time.Unix(0, 0).UTC() {
				t.Errorf("Empty value 'Download/Date' when testing URL %s", url)
			}
			if time.Since(dl.Date).Hours() > 96 && dl.Downloads == 0 {
				// Only fail for downloads that are reasonably old. Some may actually have 0 downloads
				t.Errorf("Empty value 'Download/Downloads' when testing URL %s", url)
			}
			if dl.GameVersion == "" {
				t.Errorf("Empty value 'Download/GameVersion' when testing URL %s", url)
			}
			if dl.Name == "" {
				t.Errorf("Empty value 'Download/Name' when testing URL %s", url)
			}
			if dl.ReleaseType == "" {
				t.Errorf("Empty value 'Download/ReleaseType' when testing URL %s", url)
			}
			if dl.URL == nil || dl.URL.Host == "" {
				t.Errorf("Empty value 'Download/ReleaseType' when testing URL %s", url)
			}
		}
	}
	if len(results.Authors) == 0 {
		t.Errorf("Empty list 'Authors' when testing URL %s", url)

		for _,a := range results.Authors {
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
		}
	}
	if len(results.Screenshots) == 0 {
		t.Errorf("Empty list 'Screenshots' when testing URL %s", url)

		for _,s := range results.Screenshots {

			if s.URL == nil || s.URL.Host == "" {
				t.Errorf("Empty value 'Screenshot/URL' when testing URL %s", url)
			}

			if s.ThumbnailURL != nil {
				// Thumbnail URL is not filled by the mods.curse.com parser...
				t.Errorf("'How on earch did that get here?' FILLED value 'Screenshot/ThumbnailURL' when testing URL %s", url)
			}
		}
	}
	if len(results.Categories) == 0 {
		t.Errorf("Empty list 'Categories' when testing URL %s", url)

		for _,c := range results.Categories {

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

	if results.Title == "" {
		t.Errorf("Empty value 'Title' when testing URL %s", url)
	}
	if results.License == "" {
		t.Errorf("Empty value 'License' when testing URL %s", url)
	}
	if results.Game == "" {
		t.Errorf("Empty value 'Game' when testing URL %s", url)
	}
	if results.GameURL.Host == "" {
		t.Errorf("Empty value 'GameURL' when testing URL %s", url)
	}
	if results.CurseforgeURL.Host == "" {
		t.Errorf("Empty value 'CurseforgeURL' when testing URL %s", url)
	}

	// The donation URL may actually be empty for some projects..
	if expectDonationURL {
		if results.DontationURL.Host == "" {
			t.Errorf("Empty value 'DontationURL' when testing URL %s", url)
		}
	}

	if results.Favorites == 0 {
		t.Errorf("Empty value 'Favorites' when testing URL %s", url)
	}
	if results.Likes == 0 {
		t.Errorf("Empty value 'Likes' when testing URL %s", url)
	}
	if results.AvgDownloads == 0 {
		t.Errorf("Empty value 'AvgDownloads' when testing URL %s", url)
	}
	if results.TotalDownloads == 0 {
		t.Errorf("Empty value 'TotalDownloads' when testing URL %s", url)
	}
	if results.AvgDownloadsTimeframe == "" {
		t.Errorf("Empty value 'AvgDownloadsTimeframe' when testing URL %s", url)
	}

	if results.Created == time.Unix(0, 0).UTC() {
		t.Errorf("Empty value 'Created' when testing URL %s", url)
	}
	if results.Updated == time.Unix(0, 0).UTC() {
		t.Errorf("Empty value 'Updated' when testing URL %s", url)
	}

}
