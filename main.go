package main

import (
	"fmt"
	"strings"

	rmp "github.com/Legitzx/ScheduleScraper/rate-my-professor"
	"github.com/Legitzx/ScheduleScraper/schools/pcc"
	arg "github.com/alexflint/go-arg"
)

func main() {
	var args struct {
		RunRMP     bool
		RunSchools string
	}
	arg.MustParse(&args)

	if args.RunRMP {
		fmt.Println("Scraping Rate My Professor")

		pasadenaCityCollege := 2649
		elCaminoCollege := 1403
		for _, id := range []int{pasadenaCityCollege, elCaminoCollege} {
			fmt.Printf("Running for school %d\n", id)
			professors := rmp.ScrapeRateMyProfessor(id)

			for _, professor := range professors {
				fmt.Printf("Professor: %s %s %s (%s)\n", professor.FirstName, professor.MiddleName, professor.LastName, professor.OverallRating)
			}
			fmt.Printf("Total: %d\n", len(professors))
		}
	}

	if strings.Contains(args.RunSchools, "pcc") {
		pcc.Run()
	}
}
