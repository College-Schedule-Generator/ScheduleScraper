package rmp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type FetchInfo struct {
	Base     string
	SchoolID int
	Page     int
}

type PageResult struct {
	Professors []struct {
		TDept           string `json:"tDept"`
		TSid            string `json:"tSid"`
		InstitutionName string `json:"institution_name"`
		TFname          string `json:"tFname"`
		TMiddlename     string `json:"tMiddlename"`
		TLname          string `json:"tLname"`
		Tid             int    `json:"tid"`
		TNumRatings     int    `json:"tNumRatings"`
		RatingClass     string `json:"rating_class"`
		ContentType     string `json:"contentType"`
		CategoryType    string `json:"categoryType"`
		OverallRating   string `json:"overall_rating"`
	} `json:"professors"`
	SearchResultsTotal int    `json:"searchResultsTotal"`
	Remaining          int    `json:"remaining"`
	Type               string `json:"type"`
}

type ProfessorType struct {
	Department      string `json:"department"`
	SchoolID        string `json:"schoolId"`
	InstitutionName string `json:"institutionName"`
	FirstName       string `json:"firstName"`
	MiddleName      string `json:"middleName"`
	LastName        string `json:"lastName"`
	id              int    `json:"id"`
	TotalRatings    int    `json:"totalRatings"`
	RatingsClass    string `json:"ratingsClass"`
	ContentType     string `json:"contentType"`
	CategoryType    string `json:"categoryType"`
	OverallRating   string `json:"overallRating"`
}

type ProfessorExport struct {
	Timestamp  int64           `json:"timestamp"`
	School     string          `json:"schoolId"`
	Professors []ProfessorType `json:"professors"`
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func fetchPage(info FetchInfo) PageResult {
	cacheDir := "./cache"
	cacheFile := filepath.Join(cacheDir, "rmp", fmt.Sprintf("school-%d", info.SchoolID), fmt.Sprintf("page-%d.json", info.Page))
	err := os.MkdirAll(filepath.Dir(cacheFile), 0755)
	handle(err)

	content, err := ioutil.ReadFile(cacheFile)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("=> Fetching and caching page %d\n", info.Page)
		url := fmt.Sprintf("%s/filter/professor/?&page=%d&filter=teacherlastname_sort_s+asc&query=*%%3A*&queryoption=TEACHER&queryBy=schoolId&sid=%d", info.Base, info.Page, info.SchoolID)
		resp, err := http.Get(url)
		handle(err)
		defer resp.Body.Close()

		getContent, err := ioutil.ReadAll(resp.Body)
		handle(err)

		err = ioutil.WriteFile(cacheFile, getContent, 0644)
		handle(err)

		content = getContent
	} else if err != nil {
		handle(err)
	} else {
		fmt.Printf("=> Using cached page %d\n", info.Page)
	}

	var obj PageResult
	err = json.Unmarshal(content, &obj)
	handle(err)

	return obj
}

func ScrapeRateMyProfessor(schoolId int) []ProfessorType {
	page := 1

	var professors []ProfessorType
	for {
		result := fetchPage(FetchInfo{
			Base:     "https://ratemyprofessors.com",
			SchoolID: schoolId,
			Page:     page,
		})
		if len(result.Professors) == 0 {
			break
		}

		for _, prof := range result.Professors {
			professors = append(professors, ProfessorType{
				Department:      prof.TDept,
				SchoolID:        prof.TSid,
				InstitutionName: prof.InstitutionName,
				FirstName:       prof.TFname,
				MiddleName:      prof.TMiddlename,
				LastName:        prof.TLname,
				id:              prof.Tid,
				TotalRatings:    prof.TNumRatings,
				RatingsClass:    prof.RatingClass,
				ContentType:     prof.ContentType,
				CategoryType:    prof.CategoryType,
				OverallRating:   prof.OverallRating,
			})
		}

		page++
	}

	return professors
}
