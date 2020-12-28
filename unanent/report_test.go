package unanet

import (
	"fmt"
	"testing"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

func initReportTestConnection(t *testing.T) *teamworkapi.Connection {
	conn, err := teamworkapi.NewConnectionFromJSON("./testdata/apiConfig.json")
	if err != nil {
		t.Fatalf(err.Error())
	}

	return conn
}

func TestNewReport(t *testing.T) {

	conn := initReportTestConnection(t)

	r, err := NewReport(conn)
	if err != nil {
		t.Error(err.Error())
	}

	if r.Connection == nil {
		t.Error("teamworkapi.Connection is nil")
	}

	c, err := conn.GetCompanies()
	if err != nil {
		t.Error(err.Error())
	}

	if len(c) < 1 {
		t.Error("connection test failed to return any companies")
	}
}

func TestLoadConfig(t *testing.T) {

	conn := initReportTestConnection(t)

	r, err := NewReport(conn)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = r.LoadConfig("./testdata/peopleTimeDetails.json")
	if err != nil {
		t.Errorf(err.Error())
	}

	if r.Name != "PeopleTimeDetails" {
		t.Errorf("expected report name (PeopleTimeDetails) but got (%s)", r.Name)
	}
}

func TestVerifyColumns(t *testing.T) {

	conn := initReportTestConnection(t)

	r, err := NewReport(conn)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = r.LoadConfig("./testdata/peopleTimeDetails.json")
	if err != nil {
		t.Errorf(err.Error())
	}

	testCols := []string{
		"PersonOrganization",
		"Person",
		"ProjectOrganization",
		"ProjectCode",
		"TaskNumber",
		"Task",
		"LaborCategory",
		"Location",
		"ProjectType",
		"PayCode",
		"Reference",
		"Boo",
		"ADJPostedDate",
		"FinancialPostedDate",
		"Yah",
	}

	var tests = []struct {
		cols []string
		err  bool
		want string
	}{
		{[]string{"test1", "test2", "test3"}, true, fmt.Sprintf("required number of columns (%d) not present - found 3", len(r.Fields))},
		{r.Fields, false, ""},
		{testCols, true, "required columns not found: Date, Hours"},
	}

	for _, v := range tests {
		err := r.VerifyColumns(v.cols)
		if err != nil {
			if !v.err {
				t.Error(err.Error())
			} else {
				if err.Error() != v.want {
					t.Errorf("expected error msg (%s) but got (%s)", v.want, err.Error())
				}
			}
		}
	}
}

func TestConvertUnanetDate(t *testing.T) {

	tests := []struct {
		input string
		want  string
	}{
		{"10/05/2020", "20201005"},
		{"01/22/2022", "20220122"},
		{"Bad Mojo", ""},
		{"5/2/20", ""},
	}

	for _, test := range tests {
		d, _ := ConvertUnanetDate(test.input)

		if d != test.want {
			t.Errorf("Expected %s but got %s for date string %s.", test.want, d, test.input)
		}
	}
}
