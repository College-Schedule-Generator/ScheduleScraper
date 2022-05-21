package pcc

import (
	"context"
	"fmt"
	"log"

	"github.com/Legitzx/ScheduleScraper/db"
	"github.com/PuerkitoBio/goquery"

	"net/http"
	"strconv"
	"strings"
	"unicode"
)

type MeetingTime struct {
	monday    bool
	tuesday   bool
	wednesday bool
	thursday  bool
	friday    bool
	saturday  bool
	sunday    bool
	startTime Time
	endTime   Time
}

type MeetingTimeExport struct {
	Monday    bool       `json:"Monday"`
	Tuesday   bool       `json:"Tuesday"`
	Wednesday bool       `json:"Wednesday"`
	Thursday  bool       `json:"Thursday"`
	Friday    bool       `json:"Friday"`
	Saturday  bool       `json:"Saturday"`
	Sunday    bool       `json:"Sunday"`
	StartTime TimeExport `json:"StartTime"`
	EndTime   TimeExport `json:"EndTime"`
}

// 24-hour format
type Time struct {
	hour   int
	minute int
}

type TimeExport struct {
	Hour   int `json:"Hour"`
	Minute int `json:"Minute"`
}

type DateRange struct {
	startMonth int
	startDay   int
	endMonth   int
	endDay     int
}

type DateRangeExport struct {
	StartMonth int `json:"StartMonth"`
	StartDay   int `json:"StartDay"`
	EndMonth   int `json:"EndMonth"`
	EndDay     int `json:"EndDay"`
}

type Class struct {
	courseName          string
	classID             string
	instructor          string
	availability        string
	instructionalMethod string
	meetingTimes        []MeetingTime
	date                DateRange
}

type ClassExport struct {
	CourseName          string              `json:"courseName"`
	ClassID             string              `json:"classID"`
	Instructor          string              `json:"instructor"`
	Availability        string              `json:"availability"`
	InstructionalMethod string              `json:"instructionalMethod"`
	MeetingTimes        []MeetingTimeExport `json:"meetingTimes"`
	Date                DateRangeExport     `json:"date"`
}

var globalMeetingIndex int = 0
var globalOpenVar string = ""

func Run() {
	classes := make([]Class, 1106) // 1106 is the amount of classes
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

	//analyzeArray(classes)

	saveToMongo(classes)
}

func analyzeLine(line string, courseName string, class *Class) bool {
	rawLine := line
	line = strings.ToLower(line)

	// Check if we are at the beggining of a CLASS
	if strings.Contains(line, "open") || strings.Contains(line, "closed") || strings.Contains(line, "restricted") ||
		strings.Contains(line, "see instructor") || strings.Contains(line, "waitlisted") || strings.Contains(line, "permission of dean") ||
		strings.Contains(line, "audition required") { // missing one case, idk dont want to look thru all the classes
		globalOpenVar = line
		return true
	}

	// instructional method
	switch line {
	case "l":
		class.instructionalMethod = "IP"
		return false
	case "ll":
		class.instructionalMethod = "IP"
		return false
	case "od":
		class.instructionalMethod = "FO"
		return false
	case "hy":
		class.instructionalMethod = "HY"
		return false
	}

	// class id/crn
	if _, err := strconv.Atoi(line); err == nil {
		if len(line) == 5 {
			// ok its a crn
			class.classID = line

			return false
		}
	}

	// availability
	if strings.TrimSpace(globalOpenVar) == "open" || strings.TrimSpace(globalOpenVar) == "closed" || strings.TrimSpace(globalOpenVar) == "waitlisted" {
		class.availability = strings.TrimSpace(globalOpenVar)
		globalOpenVar = ""
	}

	// course name
	if strings.TrimSpace(class.courseName) == "" {
		class.courseName = courseName
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

			if isPMStart && startHour != 12 {
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

	if isLetter(rawLine) && len(line) > 3 && !strings.Contains(line, "assignment") && !strings.Contains(line, "arranged") && !strings.Contains(line, "lecture") && !strings.Contains(line, "lab") && !strings.Contains(line, "online") {
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

func isLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' && r != '.' && r != 'I' {
			return false
		}
	}
	return true
}

func analyzeArray(classes []Class) {
	for i := 0; i < len(classes); i++ {
		fmt.Println("----------------------")
		fmt.Println(classes[i].courseName)
		fmt.Println(classes[i].classID)
		fmt.Println(classes[i].instructor)
		fmt.Println(classes[i].instructionalMethod)
		fmt.Println(classes[i].availability)
		fmt.Println(classes[i].date.startMonth)
		fmt.Println(classes[i].date.startDay)
		fmt.Println(classes[i].date.endMonth)
		fmt.Println(classes[i].date.endDay)

		for x := 0; x < len(classes[i].meetingTimes); x++ {
			fmt.Println("Day: " + strconv.Itoa(x+1))
			fmt.Println(classes[i].meetingTimes[x].monday)
			fmt.Println(classes[i].meetingTimes[x].tuesday)
			fmt.Println(classes[i].meetingTimes[x].wednesday)
			fmt.Println(classes[i].meetingTimes[x].thursday)
			fmt.Println(classes[i].meetingTimes[x].friday)
			fmt.Println(classes[i].meetingTimes[x].saturday)
			fmt.Println(classes[i].meetingTimes[x].sunday)
			fmt.Println("Meeting time: " + strconv.Itoa(classes[i].meetingTimes[x].startTime.hour) + ":" + strconv.Itoa(classes[i].meetingTimes[x].startTime.minute) + " - " + strconv.Itoa(classes[i].meetingTimes[x].endTime.hour) + ":" + strconv.Itoa(classes[i].meetingTimes[x].endTime.minute))
		}
		fmt.Println("----------------------")
	}
}

func saveToMongo(classes []Class) {
	collection, err := db.GetDBCollection()

	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < len(classes); i++ {
		count := 0

		for x := 0; x < len(classes[i].meetingTimes); x++ {
			if classes[i].meetingTimes[x].endTime.hour != 0 {
				count++
			}
		}

		// there is times
		if count > 0 {
			meetingTimeExport := make([]MeetingTimeExport, count)

			for x := 0; x < count; x++ {
				meetingTimeExport[x].Monday = classes[i].meetingTimes[x].monday
				meetingTimeExport[x].Tuesday = classes[i].meetingTimes[x].tuesday
				meetingTimeExport[x].Wednesday = classes[i].meetingTimes[x].wednesday
				meetingTimeExport[x].Thursday = classes[i].meetingTimes[x].thursday
				meetingTimeExport[x].Friday = classes[i].meetingTimes[x].friday
				meetingTimeExport[x].Saturday = classes[i].meetingTimes[x].saturday
				meetingTimeExport[x].Sunday = classes[i].meetingTimes[x].sunday

				meetingTimeExport[x].StartTime.Hour = classes[i].meetingTimes[x].startTime.hour
				meetingTimeExport[x].StartTime.Minute = classes[i].meetingTimes[x].startTime.minute

				meetingTimeExport[x].EndTime.Hour = classes[i].meetingTimes[x].endTime.hour
				meetingTimeExport[x].EndTime.Minute = classes[i].meetingTimes[x].endTime.minute
			}

			newClass := ClassExport{classes[i].courseName, classes[i].classID, classes[i].instructor,
				classes[i].availability, classes[i].instructionalMethod, meetingTimeExport, DateRangeExport{classes[i].date.startMonth, classes[i].date.startDay, classes[i].date.endMonth, classes[i].date.endDay}}
			_, err := collection.InsertOne(context.TODO(), newClass)
			if err != nil {
				fmt.Println(err)
			}
		} else { // no meeting times
			newClass := ClassExport{classes[i].courseName, classes[i].classID, classes[i].instructor,
				classes[i].availability, classes[i].instructionalMethod, nil, DateRangeExport{classes[i].date.startMonth, classes[i].date.startDay, classes[i].date.endMonth, classes[i].date.endDay}}
			_, err := collection.InsertOne(context.TODO(), newClass)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

/* TODO
- last crn is not showing up
 - one instructor not being read 1102/1103 intructors
 - last course not showing up
*/
