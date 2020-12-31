package unanet

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"time"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

// TimeDetailsReport models a specific type of Unanet report - the Person Time
// Details report.
type TimeDetailsReport struct {
	Report    *Report
	Entries   []*teamworkapi.TimeEntry
	StartDate time.Time
	EndDate   time.Time
}

// NewTimeDetailsReport creates a new TimeDetailsReport and initializes the
// underlying parent Report object with the specific teamworkapi.Connection.
func NewTimeDetailsReport(conn *teamworkapi.Connection) (*TimeDetailsReport, error) {

	t := new(TimeDetailsReport)

	r, err := NewReport(conn)
	if err != nil {
		return nil, err
	}

	t.Report = r

	return t, nil
}

// ParseTimeDetailsReport converts Unanet time report into an array of TimeEntry
// objects and updates the TimeDetailsReport with applicable metadata, including
// the TimeEntry array, and start/end date of report.
func (r *TimeDetailsReport) ParseTimeDetailsReport(path string) ([]*teamworkapi.TimeEntry, error) {

	users := make(map[string]*teamworkapi.Person)

	m := r.Report.FieldIndex // maps column names to respective index (0, 1, 2...)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r.Report.Filename = path

	rdr := csv.NewReader(f)

	// read the first row to get column names
	cols, err := rdr.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read column names from file %s", f.Name())
	}

	// verify required fields are present
	err = r.Report.VerifyColumns(cols)
	if err != nil {
		return nil, err
	}

	lines, err := rdr.ReadAll()
	if err != nil {
		return nil, err
	}

	entries := make([]*teamworkapi.TimeEntry, len(lines))

	var startDate, endDate time.Time

	for i, line := range lines {

		// if we have not seen this person yet, initialize a new Person object
		if _, ok := users[line[m["Person"]]]; !ok {

			// verify person is found in Teamwork
			p, err := r.Report.Connection.GetPersonByID(r.Report.UserMappings[line[m["Person"]]])
			if err != nil {
				return nil, err
			}

			users[line[m["Person"]]] = p
		}

		// verify the Unanet Project Code has been mapped to a Teamwork Project ID
		if _, ok := r.Report.ProjectMappings[line[m["ProjectCode"]]]; !ok {
			return nil, fmt.Errorf("no Project mapping found for Unanet project %s", line[m["ProjectCode"]])
		}
		
		// verify the Unanet Task Number has been mapped to a Teamwork Task ID
		if _, ok := r.Report.TaskMappings[line[m["TaskNumber"]]]; !ok {
			return nil, fmt.Errorf("no Task mapping found for Unanet task %s", line[m["TaskNumber"]])
		}

		entryDate, err := time.Parse(UnanetTimeShort, line[m["Date"]])
		if err != nil {
			return nil, err
		}

		// if first iteration, set start/end dates for this report, otherwise,
		// update start/end date as applicable
		if i == 0 {
			startDate = entryDate
			endDate = entryDate
		} else {
			if entryDate.Before(startDate) {
				startDate = entryDate
			}

			if entryDate.After(endDate) {
				endDate = entryDate
			}
		}

		// create a new time entry for the current person
		entry := new(teamworkapi.TimeEntry)
		entry.PersonID = users[line[m["Person"]]].ID
		entry.Date = entryDate.Format(TeamworkTimeShort)
		entry.Description = "Imported from Unanet Time Details report."
		entry.Hours = line[m["Hours"]]
		entry.Minutes = "0"
		entry.ProjectID = r.Report.ProjectMappings[line[m["ProjectCode"]]]
		entry.TaskID = r.Report.TaskMappings[line[m["TaskNumber"]]]

		entries[i] = entry
	}

	r.StartDate = startDate
	r.EndDate = endDate
	r.Entries = entries

	err = f.Close()
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// UploadTimeEntries uploads the time entries is the TimeDetailsReport to
// Teamwork.  A new file is created that includes all the original data from the
// Unanet report, plus a new column TeamworkTimeEntryID that maps the Unanet
// time log to a Teamwork Time Entry.  The new file is saved with the naming
// convention, time-details-report_YYYYMMDD-YYYYMMDD.csv.
func (r *TimeDetailsReport) UploadTimeEntries() error {

	errorBuffer := ""

	originalFile := r.Report.Filename
	newFileName := fmt.Sprintf("time-details-report_%s-%s.csv", r.StartDate.Format(TeamworkTimeShort), r.EndDate.Format(TeamworkTimeShort))

	oFile, err := os.Open(originalFile)
	if err != nil {
		return err
	}

	nFile, err := os.Create(path.Join(path.Dir(originalFile), newFileName))
	if err != nil {
		return err
	}

	r.Report.Filename = newFileName

	rdr := csv.NewReader(oFile)
	wrtr := csv.NewWriter(nFile)

	// read the first row to get column names
	cols, err := rdr.Read()
	if err != nil {
		return fmt.Errorf("failed to read column names from file %s", oFile.Name())
	}

	// add a column for the Teamwork Time Entry ID
	cols = append(cols, "TeamworkTimeEntryID")

	// write columns to new file
	err = wrtr.Write(cols)
	if err != nil {
		return err
	}

	for _, e := range r.Entries {

		toAppend := ""

		e.Description = fmt.Sprintf("Imported from Unanet People Time Details report, %s", path.Base(r.Report.Filename))

		cols, err = rdr.Read()
		if err != nil {
			if errorBuffer != "" {
				errorBuffer += "; "
			}
			errorBuffer += err.Error()
			continue
		}

		id, err := r.Report.Connection.PostTimeEntry(e)
		if err != nil {
			toAppend = err.Error()

			if errorBuffer != "" {
				errorBuffer += "; "
			}
			errorBuffer += err.Error()
		} else {
			toAppend = id
		}

		cols = append(cols, toAppend)

		err = wrtr.Write(cols)
		if err != nil {
			if errorBuffer != "" {
				errorBuffer += "; "
			}
			errorBuffer += err.Error()
		}
	}

	err = oFile.Close()
	if err != nil {
		if errorBuffer != "" {
			errorBuffer += "; "
		}
		errorBuffer += err.Error()
	}

	wrtr.Flush()

	err = nFile.Close()
	if err != nil {
		if errorBuffer != "" {
			errorBuffer += "; "
		}
		errorBuffer += err.Error()
	}

	if errorBuffer != "" {
		return fmt.Errorf("one or more errors occurred: %s", errorBuffer)
	}

	return nil
}
