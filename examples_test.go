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
	"log"
	"net/url"

	curse "github.com/founderio/curse-parser"
)

// This example illustrates fetching & parsing of a Curse page
func Example_curse() {
	projectURL := "https://mods.curse.com/mc-mods/minecraft/238424-taam"
	// Fetch the page content
	resp, err := curse.FetchPage(projectURL)
	if err != nil {
		log.Fatal("error fetching page:", err)
	}

	// Parse the page content
	// (URL is required for parsing some links or image URLs)
	results, err := curse.ParseCurse(projectURL, resp)
	if err != nil {
		log.Fatal("error parsing page:", err)
	}

	log.Println("The Curse project", results.Title, "was created at", results.Created)
	log.Println("Most recent update:", results.Updated)
	log.Println("View the CurseForge page at", results.CurseforgeURL.String())
	log.Println("Files:")

	for _, v := range results.Downloads {
		log.Println(v.Name, " - ", v.Date, " - for ", v.GameVersion)
	}

}

// This example illustrates fetching & parsing of a CurseForge page
func Example_curseForge() {
	// CurseForge parse expects URLs
	projectURL, err := url.Parse("https://minecraft.curseforge.com/projects/taam")
	if err != nil {
		log.Fatal("error parsing URL:", err)
	}

	log.Println("Fetching Pages...")

	// This will parse the overview page and the files page using default options.
	// That means, the file parse will parse ALL file pages! Use curse.CFOptionFilesNoPagination to prevent that.
	results, err := curse.FetchCurseForge(projectURL, curse.CFSectionOverview|curse.CFSectionFiles, curse.CFOptionNone)
	if err != nil {
		log.Fatal("error fetching CurseForge project:", err)
	}

	log.Println("The CurseForge project", results.Title, "was created at", results.Created)
	log.Println("Most recent update:", results.Updated)
	log.Println("Files:")

	for _, v := range results.Downloads {
		log.Println(v.Name, " - ", v.Date, " - for ", v.GameVersion)
	}

}
