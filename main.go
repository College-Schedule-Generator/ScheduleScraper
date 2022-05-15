package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

type MeetingTime struct {
	monday bool
	tuesday bool
	wednesday bool
	thursday bool
	friday bool
	saturday bool
	sunday bool
	startTime Time
	endTime Time
}

// 24-hour format
type Time struct {
	hour int
	minute int
}

type DateRange struct {
	startMonth int
	startDay int
	endMonth int
	endDay int
}

type Class struct {
	courseName string
	crn string
	instructor string
	instructorRating float32
	open bool
	online bool
	meetingTimes[] MeetingTime
	date DateRange
}

var globalMeetingIndex int = 0

func main() {
	classes := make([]Class, 1103)
	classesIndex := 0
	currentClass := Class{}

	webPage := "http://127.0.0.1:5500/schedule.html"
	resp, err := http.Get(webPage)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	courseName := ""
	doc.Find("tr").Each(func(index int, item *goquery.Selection) {
		courseNameRaw := item.Find(".crn_header").Text()
		courseNameArr := strings.Split(courseNameRaw, " ")
		check := false

		if len(courseNameArr) > 1 {
			courseName = courseNameArr[0] + " " + courseNameArr[1]
			fmt.Println("\n" + courseName)
		}


		item.Find(".default1").Each(func(i int, item2 *goquery.Selection) {
			it := strings.TrimSpace(item2.Text())

			if it != "" {
				fmt.Println("1[" + it + "]")

				if analyzeLine(it, courseName, &currentClass) {
					classes[classesIndex] = currentClass
					classesIndex++
					currentClass = Class{}
				}
			}

			check = true
		})

		item.Find(".default2").Each(func(i int, item2 *goquery.Selection) {
			it := strings.TrimSpace(item2.Text())

			if it != "" {
				fmt.Println("2[" + it + "]")

				if analyzeLine(it, courseName, &currentClass) {
					classes[classesIndex] = currentClass
					classesIndex++
					currentClass = Class{}
				}
			}

			check = true
		})

		if check {
			fmt.Println("\n")
		}
	})

	analyzeArray(classes)
}

func analyzeLine(line string, courseName string, class *Class) bool {
	rawLine := line
	line = strings.ToLower(line)

	// Check if we are at the beggining of a CLASS
	if strings.Contains(line, "open") || strings.Contains(line, "closed") || strings.Contains(line, "restricted") ||
		strings.Contains(line, "see instructor") || strings.Contains(line, "waitlisted") || strings.Contains(line, "permission of dean") ||
		strings.Contains(line, "audition required"){ // missing one case, idk dont want to look thru all the classes
		if strings.Contains(line, "open") {
			class.open = true
		}
		return true
	}

	// course name
	if strings.TrimSpace(class.courseName) == "" {
		class.courseName = courseName
	}

	// crn
	if _, err := strconv.Atoi(line); err == nil {
		if len(line) == 5 {
			// ok its a crn
			class.crn = line

			return false
		}
	}

	// check if online
	if strings.Contains(line, "online") {
		class.online = true
		return false
	}

	// DATES RANGE contain / - /
	if strings.Contains(line, "/") && strings.Contains(line, "-") && class.date.startMonth == 0 {
		// parse date
		startMonth, err1 := strconv.Atoi(line[:2])
		startDay, err2 := strconv.Atoi(line[3:5])

		endMonth, err3 := strconv.Atoi(line[6:8])
		endDay, err4 := strconv.Atoi(line[9:])

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			fmt.Println(err1)
			fmt.Println(err2)
			fmt.Println(err3)
			fmt.Println(err4)
		}

		class.date.startMonth = startMonth
		class.date.startDay = startDay
		class.date.endMonth = endMonth
		class.date.endDay = endDay

		return false
	}


	// Meeting TIMES : - :
	if strings.Contains(line, ":") && strings.Contains(line, "-") {
		// parse time
		if len(class.meetingTimes) > 0 {
			startHour, err1 := strconv.Atoi(line[:2])
			startMinute, err2 := strconv.Atoi(line[3:5])

			isPMRawStart := line[5:7]
			isPMStart := false

			endHour, err3 := strconv.Atoi(line[10:12])
			endMinute, err4 := strconv.Atoi(line[13:15])

			isPMRawEnd := line[15:]
			isPMEnd := false

			if isPMRawStart == "pm" {
				isPMStart = true
			}

			if isPMRawEnd == "pm" {
				isPMEnd = true
			}

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				fmt.Println(err1)
				fmt.Println(err2)
				fmt.Println(err3)
				fmt.Println(err4)
			}

			if isPMStart && startHour != 12{
				startHour += 12
			}

			if isPMEnd && endHour != 12 {
				endHour += 12
			}

			class.meetingTimes[globalMeetingIndex].startTime.hour = startHour
			class.meetingTimes[globalMeetingIndex].startTime.minute = startMinute

			class.meetingTimes[globalMeetingIndex].endTime.hour = endHour
			class.meetingTimes[globalMeetingIndex].endTime.minute = endMinute
		}

		return false
	}


	// Meeting DAYS
	if len(line) <= 2 {
		if strings.Contains(line, "m") || strings.Contains(line, "t") || strings.Contains(line, "w") ||
			strings.Contains(line, "th") || strings.Contains(line, "f") || strings.Contains(line, "s") || strings.Contains(line, "su") {
			if len(class.meetingTimes) == 0 {
				class.meetingTimes = make([]MeetingTime, 10) // yes, one of the classes actually has 10 meeting times...
				globalMeetingIndex = 0
				class.meetingTimes[0] = MeetingTime{}
			}

			if len(class.meetingTimes) > 0 {
				if class.meetingTimes[globalMeetingIndex].endTime.hour != 0 {
					globalMeetingIndex++
					class.meetingTimes[globalMeetingIndex] = MeetingTime{}
				}

				switch line {
					case "m":
						// monday
						class.meetingTimes[globalMeetingIndex].monday = true
					case "t":
						class.meetingTimes[globalMeetingIndex].tuesday = true
					case "w":
						class.meetingTimes[globalMeetingIndex].wednesday = true
					case "th":
						class.meetingTimes[globalMeetingIndex].thursday = true
					case "f":
						class.meetingTimes[globalMeetingIndex].friday = true
					case "s":
						class.meetingTimes[globalMeetingIndex].saturday = true
					case "su":
						class.meetingTimes[globalMeetingIndex].sunday = true
				}
			}

			return false
		}
	}

	// instructor. Two words (only letters). Only thing that can conflict with this is the location. Location is usually in all CAPS, so if we get original string and check if its uppercase, then its not instructor.
	// check if contains numbers
	if line == "staff" {
		class.instructor = line
		return false
	}

	if IsLetter(rawLine) && len(line) > 3 && !strings.Contains(line, "assignment") && !strings.Contains(line, "arranged") && !strings.Contains(line, "lecture") && !strings.Contains(line, "lab") {
		count := 0
		for i := 0; i < len(line); i++ {
			if line[i] != rawLine[i] {
				count++
			}
		}

		// check if contains caps
		if count <= 5 {
			// check if is two words
			if len(strings.Split(line, " ")) <= 5 {
				class.instructor = line
			}
		}
	}

	return false
}

func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' && r != '.' && r != 'I'{
			return false
		}
	}
	return true
}

func analyzeArray(classes []Class) {
	for i := 0; i < len(classes); i++ {
		fmt.Println("----------------------")
		fmt.Println(classes[i].courseName)
		fmt.Println(classes[i].crn)
		fmt.Println(classes[i].instructor)
		fmt.Println(classes[i].online)
		fmt.Println(classes[i].date.startMonth)
		fmt.Println(classes[i].date.startDay)
		fmt.Println(classes[i].date.endMonth)
		fmt.Println(classes[i].date.endDay)

		for x := 0; x < len(classes[i].meetingTimes); x++ {
			fmt.Println("Day: " + strconv.Itoa(x + 1))
			fmt.Println(classes[i].meetingTimes[x].monday)
			fmt.Println(classes[i].meetingTimes[x].tuesday)
			fmt.Println(classes[i].meetingTimes[x].wednesday)
			fmt.Println(classes[i].meetingTimes[x].thursday)
			fmt.Println(classes[i].meetingTimes[x].friday)
			fmt.Println(classes[i].meetingTimes[x].saturday)
			fmt.Println(classes[i].meetingTimes[x].sunday)
			fmt.Println("Meeting time: " + strconv.Itoa(classes[i].meetingTimes[x].startTime.hour) + ":" + strconv.Itoa(classes[i].meetingTimes[x].startTime.minute) + " - " + strconv.Itoa(classes[i].meetingTimes[x].endTime.hour) + ":" + strconv.Itoa(classes[i].meetingTimes[x].endTime.minute));
		}
		fmt.Println("----------------------")
	}
}

/* TODO
- last crn is not showing up
 - one instructor not being read 1102/1103 intructors
 - last course not showing up
 */