package unanet

import (
	"testing")


func TestNewReportFromJSON(t *testing.T) {

	_, err := NewReportFromJSON("conf/peopleTimeDetails.json")

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestConvertUnanetDate(t *testing.T) {

	tests := []struct {
		input	string
		want	string
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