package main

import (
	"context"
	"fmt"
	"github.com/Legitzx/ScheduleScraper/db"
	"strconv"
	"strings"
	"time"

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
		fmt.Println("Scraping and Inserting Rate My Professor data into database")

		pasadenaCityCollege := 2649
		elCaminoCollege := 1403
		for _, id := range []int{pasadenaCityCollege, elCaminoCollege} {
			fmt.Printf("Running for school %d\n", id)
			professors := rmp.ScrapeRateMyProfessor(id)

			collection, err := db.GetDBCollection("Professors")
			if err != nil {
				fmt.Println(err)
				return
			}

			professorExport := rmp.ProfessorExport{Timestamp: time.Now().Unix(), SchoolId: strconv.Itoa(id), Professors: professors}

			_, err = collection.InsertOne(context.TODO(), professorExport)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	if strings.Contains(args.RunSchools, "pcc") {
		pcc.Run()
	}
}
