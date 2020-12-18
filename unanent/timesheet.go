package unanet

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	teamworkapi "github.com/Foxtrot-Division/teamworkAPI"
)

// Users provides a mapping of a Unanet user to a Teamwork user.  The map
// is indexed by the lastname, firstname string found in Unanet time reports.
// is indexed by the lastname, firstname string found in Unanet time reports.
type Users map[string]teamworkapi.Person

// ParsePeopleTimeReport converts Unanet time report into a TimeEntry object.
func (r Report) ParsePeopleTimeReport(path string) ([]*teamworkapi.TimeEntry, error) {

	users := make(map[string]*teamworkapi.Person)

	m := r.FieldIndex

	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	rdr := csv.NewReader(f)

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


		if _, ok := users[line[m["Person"]]]; !ok {
			name := strings.Split(line[m["Person"]], ",")

			users[line[m["Person"]]] = new(teamworkapi.Person)

			users[line[m["Person"]]].LastName = name[0]
			users[line[m["Person"]]].FirstName = name[1]
			users[line[m["Person"]]].ID = r.UserMappings[line[m["Person"]]]
			users[line[m["Person"]]].CompanyName = r.CompanyMappings[line[m["PersonOrganization"]]]

			fmt.Printf("[%d] Created user %s (%s)\n", i, users[line[m["Person"]]].LastName, users[line[m["Person"]]].ID) 
		}

		date, err := ConvertUnanetDate(line[m["Date"]])
		if err != nil {
			return nil, err
		}


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
func (r Report) UploadTimeEntries(entries []*teamworkapi.TimeEntry) (error) {
	
	errorBuffer := ""
	
	for _, e := range entries {

		_, err := r.UploadTimeEntry(e)

		if err != nil {
			errorBuffer += err.Error()
		}
	}

	if errorBuffer != "" {
		return fmt.Errorf(errorBuffer)
	} 
		
	return nil
}

// UploadTimeEntry uploads a single TimeEntry to Teamwork and returns the
// entry's ID.
func (r Report) UploadTimeEntry(entry *teamworkapi.TimeEntry) (string, error) {

	type response struct {
		Status 		string `json:"STATUS"`
		Message 	string `json:"MESSAGE"`
		TimeLogID 	string `json:"timeLogId"`
	}

	entryJSON := new(teamworkapi.TimeEntryJSON)
	entryJSON.Entry = *entry

	reqData, err := json.Marshal(entryJSON)
	if err != nil {
		return "", err
	}

	conn, err := teamworkapi.NewConnectionFromJSON("testdata/apiConfig.json")
	if err != nil {
		return "", err
	}

	resMsg := new(response)
	resData, err := conn.PostRequest("/tasks/" + entry.TaskID + "/time_entries", reqData)
	if err != nil {
		return "", err
	}

	fmt.Println(string(resData))
	
	err = json.Unmarshal(resData, &resMsg)
	if err != nil {
		return "", err
	}

	if resMsg.Status != "OK" {
		return "", fmt.Errorf(resMsg.Message)
	}

	return resMsg.TimeLogID, nil
}