package unanet

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

func initTimeDetailsReport(t *testing.T) *Report {

	conn, err := teamworkapi.NewConnectionFromJSON("./testdata/apiConfig.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	r, err := NewReport(conn)
	if err != nil {
		t.Errorf(err.Error())
	}

	r.LoadConfig("./testdata/timeDetails.json")

	return r
}

func TestParseTimeDetailsReport(t *testing.T) {

	r := initTimeDetailsReport(t)

	entries, err := r.ParseTimeDetailsReport("./testdata/report.csv")
	if err != nil {
		t.Errorf(err.Error())
	}

	sumHours := 0.0
	userEntries := make(map[string] *teamworkapi.Person)

	for _, v := range entries {

		h, err := strconv.ParseFloat(v.Hours, 64)
		if err != nil {
			t.Error(err.Error())
		}

		sumHours += h

		if userEntries[v.PersonID] == nil {

			// verify we've mapped the user correctly by retrieving the user details
			p, err := r.Connection.GetPersonByID(v.PersonID) 
			if err != nil {
				t.Error(err.Error())
			}
		
			if p == nil {
				t.Errorf("call to GetPersonByID(%s) returned nil", v.PersonID)
			}

			userEntries[v.PersonID] = p
		}
	}

	if sumHours < 1.0 {
		t.Errorf("expected total hours to be at least 1.0 but got (%f)", sumHours)
	}
}

func TestUploadTimeEntries(t *testing.T) {

	r := initTimeDetailsReport(t)

	f, err := os.Open("./testdata/timeDetailsTestData.json")
	defer f.Close()

	if err != nil {
		t.Fatalf(err.Error())
	}

	data := new(teamworkapi.TimeEntriesJSON)

	raw, _ := ioutil.ReadAll(f)

	err = json.Unmarshal(raw, &data)
	if err != nil {
		t.Fatalf(err.Error())
	}

	entries := data.TimeEntries

	err = r.UploadTimeEntries(entries)
	if err != nil {
		t.Errorf(err.Error())
	}
}