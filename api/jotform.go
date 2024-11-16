package jotform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func loginAndGetSessionID(loginURL, databaseID, username, password string) (string, error) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("database", databaseID)
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Extract session cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_id" { // Adjust to match actual cookie name
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("session ID not found")
}

func createPerson(apiURL, sessionID string, personData map[string]string) error {
	client := &http.Client{}
	jsonData, err := json.Marshal(personData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", fmt.Sprintf("asm_session_id=%s", sessionID))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create person: %s", resp.Status)
	}
	return nil
}

func Jotform(w http.ResponseWriter, r *http.Request) {
	// Read environment variables
	baseURL := "https://us03d.sheltermanager.com"
	loginURL := baseURL + "/login"
	databaseID := os.Getenv("SHELTER_MANAGER_DATABASE_ID")
	username := os.Getenv("SHELTER_MANAGER_USERNAME")
	password := os.Getenv("SHELTER_MANAGER_PASSWORD")

	sessionID, err := loginAndGetSessionID(loginURL, databaseID, username, password)
	if err != nil {
		fmt.Fprintf(w, "Error:\n\n%s", err)
	}

	newPersonURL := baseURL + "/person_new"
	personData := map[string]string{
		"ownertype":        "1",
		"forenames":        "Test",
		"surname":          "Testerson",
		"country":          "USA",
		"jurisdiction":     "1",
		"flags":            "",
		"gdprcontactoptin": "",
		"site":             "0",
		"a.1.7":            "No",
	}
	err = createPerson(newPersonURL, sessionID, personData)
	if err != nil {
		fmt.Fprintf(w, "Error:\n\n%s", err)
	}

	fmt.Fprint(w, "Person created successfully!")

	// Copy the request body to the response
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "Unable to read request body", http.StatusInternalServerError)
	// 	return
	// }
	// defer r.Body.Close()
	//
	// // Write the contents of the request body to the response
	// fmt.Fprintf(w, "Request body:\n\n%s", body)
}
