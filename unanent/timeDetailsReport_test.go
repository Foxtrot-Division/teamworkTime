package unanet

import (
	"fmt"
	"strconv"
	"testing"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

func initTimeDetailsReport(t *testing.T) *TimeDetailsReport {

	conn, err := teamworkapi.NewConnectionFromJSON("./testdata/apiConfig.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	tr, err := NewTimeDetailsReport(conn)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = tr.Report.LoadConfig("./testdata/timeDetailsConf.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	return tr
}

func TestParseTimeDetailsReport(t *testing.T) {

	r := initTimeDetailsReport(t)

	entries, err := r.ParseTimeDetailsReport("./testdata/report3.csv")
	if err != nil {
		t.Errorf(err.Error())
	}

	sumHours := 0.0
	userEntries := make(map[string]*teamworkapi.Person)

	for _, v := range entries {
		fmt.Printf("entry for %s on %s\n", v.PersonID, v.Date)
		h, err := strconv.ParseFloat(v.Hours, 64)
		if err != nil {
			t.Error(err.Error())
		}

		sumHours += h

		if userEntries[v.PersonID] == nil {

			// verify we've mapped the user correctly by retrieving the user details
			p, err := r.Report.Connection.GetPersonByID(v.PersonID)
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

func TestScrubTimeEntries(t *testing.T) {
	r := initTimeDetailsReport(t)

	_, err := r.ParseTimeDetailsReport("./testdata/report3.csv")
	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Printf("Original Length: %d\n", len(r.Entries))

	err = r.ScrubTimeEntries()
	if err != nil {
		t.Fatalf(err.Error())
	}


	fmt.Printf("Post-scrub length: %d\n", len(r.Entries))
}

func TestUploadTimeEntries(t *testing.T) {

	r := initTimeDetailsReport(t)

	_, err := r.ParseTimeDetailsReport("./testdata/report3.csv")
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = r.UploadTimeEntries()
	if err != nil {
		t.Errorf(err.Error())
	}
}
