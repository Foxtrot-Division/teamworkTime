package unanet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

// UnanetTimeShort represents a Unanet date in short form.
const UnanetTimeShort = "01/02/2006"

// TeamworkTimeShort represents a Teamwork date in short form.
const TeamworkTimeShort = "20060102"

// Report models a Unanet time report.  Fields represents the required columns
// in the .csv file.  FieldIndex maps each field to its respective column number
// in the .csv file.
type Report struct {
	Name            string   			`json:"reportName"`
	Fields          []string 			`json:"columns"`
	FieldIndex      map[string]int
	CompanyMappings map[string]string 	`json:"companyMappings"`
	ProjectMappings map[string]string 	`json:"projectMappings"`
	TaskMappings    map[string]string 	`json:"taskMappings"`
	UserMappings    map[string]string 	`json:"userMappings"`
	Connection		*teamworkapi.Connection
}

// NewReport creates a new Report instance and initializes the Connection to
// Teamwork API.
func NewReport(conn *teamworkapi.Connection) (*Report, error) {
	
	if conn == nil {
		return nil, fmt.Errorf("parameter (*teamworkapi.Connection) is nil")
	}

	r := new(Report)

	r.Connection = conn

	return r, nil
}

// LoadConfig configures a Report based on json file.
func (r *Report) LoadConfig(pathToFile string) (error) {

	f, err := os.Open(pathToFile)
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	r.FieldIndex = make(map[string]int)

	for i, v := range r.Fields {
		r.FieldIndex[v] = i
	}

	return nil
}

// VerifyColumns performs a sanity check to ensure the comma-separated string
// columns includes all required columns.
func (r *Report) VerifyColumns(columns []string) error {

	if len(columns) != len(r.Fields) {
		return fmt.Errorf("required number of columns (%d) not present - found %d", len(r.Fields), len(columns))
	}

	errBuff := ""

	for _, v1 := range r.Fields {
		found := false

		for _, v2 := range columns {
			if v2 == v1 {
				found = true
				break
			}
		}

		if !found {
			if errBuff != "" {
				errBuff += ", "
			}
			
			errBuff += v1
		}
	}

	if errBuff != "" {
		return fmt.Errorf("required columns not found: %s", errBuff)
	}

	return nil
}

// ConvertUnanetDate converts Unanet short format (MM/DD/YYYY) into Teamwork API
// short format (YYYYMMDD).
func ConvertUnanetDate(d string) (string, error) {
	uTime, err := time.Parse(UnanetTimeShort, d)

	if err != nil {
		return "", err
	}

	return uTime.Format(TeamworkTimeShort), nil
}
