package unanet

import (
	"fmt"
	"testing"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)


func TestParsePeopleTimeReport(t *testing.T) {

	r, err := NewReportFromJSON("conf/peopleTimeDetails.json")

	if err != nil {
		t.Errorf(err.Error())
	}

	entries, err := r.ParsePeopleTimeReport("testdata/report.csv")

	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf("Entries found: %d\n", len(entries))

	for k, v := range entries {
		fmt.Printf("%d %s %s %s %s\n", k, v.PersonID, v.TaskID, v.Date, v.Hours)
	}
}

func TestUploadTimeEntry(t *testing.T) {

	r, err := NewReportFromJSON("conf/peopleTimeDetails.json")

	if err != nil {
		t.Errorf(err.Error())
	}

	entry := teamworkapi.TimeEntry{
		PersonID: "118616",
		Date: "20201215",
		Hours: "10.5",
		Minutes: "0",
		Description: "Unanent entry.",
		IsBillable: "true",
		TaskID: "20029437",
		ProjectID: "409216",
	}

	id, err := r.UploadTimeEntry(&entry)
	if err != nil {
		t.Errorf(err.Error())
	}

	fmt.Printf("Added time log %s\n", id)
}