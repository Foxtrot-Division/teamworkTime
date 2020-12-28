package unanet

import (
	"encoding/csv"
	"fmt"
	"os"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

// ParseTimeDetailsReport converts Unanet time report into an array of TimeEntry objects.
func (r *Report) ParseTimeDetailsReport(path string) ([]*teamworkapi.TimeEntry, error) {

	users := make(map[string] *teamworkapi.Person)

	m := r.FieldIndex // maps column names to respective index (0, 1, 2...)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	rdr := csv.NewReader(f)

	// read the first row to get column names
	cols, err := rdr.Read() 
	if err != nil {
		return nil, fmt.Errorf("failed to read column names from file %s", f.Name())
	}

	// verify required fields are present
	err = r.VerifyColumns(cols)
	if err != nil {
		return nil, err
	}

	lines, err := rdr.ReadAll()
	if err != nil {
		return nil, err
	}

	entries := make([]*teamworkapi.TimeEntry, len(lines))

	for i, line := range lines {
 
		// if we have not seen this person yet, initialize a new Person object
		if _, ok := users[line[m["Person"]]]; !ok {

			p, err := r.Connection.GetPersonByID(r.UserMappings[line[m["Person"]]])
			if err != nil {
				return nil, err
			}

			users[line[m["Person"]]] = p
		}

		date, err := ConvertUnanetDate(line[m["Date"]])
		if err != nil {
			return nil, err
		}

		// create a new time entry for the current person
		entry := new(teamworkapi.TimeEntry)
		entry.PersonID = users[line[m["Person"]]].ID
		entry.Date = date
		entry.Description = "Imported from Unanet report."
		entry.Hours = line[m["Hours"]]
		entry.ProjectID = r.ProjectMappings[line[m["ProjectCode"]]]
		entry.TaskID = r.TaskMappings[line[m["TaskNumber"]]]

		entries[i] = entry
	}

	return entries, nil
}

// UploadTimeEntries uploads multiple TimeEntry instances to Teamwork.
func (r *Report) UploadTimeEntries(entries []*teamworkapi.TimeEntry) (error) {
	
	errorBuffer := ""
	
	for _, e := range entries {

		id, err := r.Connection.PostTimeEntry(e)
		if err != nil {
			if errorBuffer != "" {
				errorBuffer += "; "
			}
			errorBuffer += err.Error()
			continue
		}
	
		fmt.Printf("created time entry ID (%s)", id)
	}

	if errorBuffer != "" {
		return fmt.Errorf("one or more errors occurred: %s", errorBuffer)
	} 
		
	return nil
}
































