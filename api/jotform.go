package jotform

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func ConvertMapToValues(data map[string]string) url.Values {
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}
	return values
}

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

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "asm_session_id" {
			return cookie.Value, nil
		}
	}
	return "", fmt.Errorf("session ID not found")
}

func createPerson(apiURL, sessionID string, personData map[string]string) error {
	client := &http.Client{}
	data := ConvertMapToValues(personData)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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
	baseURL := "https://us03d.sheltermanager.com"
	loginURL := baseURL + "/login"
	databaseID := os.Getenv("SHELTER_MANAGER_DATABASE_ID")
	username := os.Getenv("SHELTER_MANAGER_USERNAME")
	password := os.Getenv("SHELTER_MANAGER_PASSWORD")

	sessionID, err := loginAndGetSessionID(loginURL, databaseID, username, password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	rawRequest := r.FormValue("rawRequest")
	if rawRequest == "" {
		http.Error(w, "rawRequest field is missing", http.StatusBadRequest)
		return
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(rawRequest), &requestData); err != nil {
		http.Error(w, "Failed to parse rawRequest JSON", http.StatusBadRequest)
		return
	}

	// Transform incoming request data to ShelterManager format
	personData := map[string]string{
		"ownertype": "1",
		"title":     "Mr.", // Default title, adjust if needed
		"forenames": requestData["q3_fullName3"].(map[string]interface{})["first"].(string),
		"surname":   requestData["q3_fullName3"].(map[string]interface{})["last"].(string),
		// Using strings.Replace to handle spaces in address properly
		"address":       strings.Replace(requestData["q4_address4"].(map[string]interface{})["addr_line1"].(string), "+", " ", -1),
		"town":          requestData["q4_address4"].(map[string]interface{})["city"].(string),
		"county":        requestData["q4_address4"].(map[string]interface{})["state"].(string),
		"postcode":      requestData["q4_address4"].(map[string]interface{})["postal"].(string),
		"country":       "USA",
		"hometelephone": requestData["q88_phoneNumber"].(map[string]interface{})["full"].(string),
		"emailaddress":  requestData["q6_email6"].(string),
		"dateofbirth":   fmt.Sprintf("%s/%s/%s", requestData["q35_dob"].(map[string]interface{})["month"], requestData["q35_dob"].(map[string]interface{})["day"], requestData["q35_dob"].(map[string]interface{})["year"]),
		// Removed idnumber field as it should come from the form
		"jurisdiction":     "1",
		"flags":            "adopter",
		"gdprcontactoptin": "",
		"site":             "0",
		"a.1.7":            "No",
	}

	// Map the ID from q86_typeA86 field
	if id, ok := requestData["q86_typeA86"].(string); ok && id != "" {
		personData["idnumber"] = id
	}

	newPersonURL := baseURL + "/person_new"
	if err := createPerson(newPersonURL, sessionID, personData); err != nil {
		http.Error(w, fmt.Sprintf("Error creating person: %s", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Person created successfully!")
	log.Printf("Transformed Request: %+v", personData)
}
