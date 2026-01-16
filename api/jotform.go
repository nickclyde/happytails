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

// Helper function to safely extract string from interface{}
func getString(data interface{}) string {
	if data == nil {
		return ""
	}
	if str, ok := data.(string); ok {
		return str
	}
	return ""
}

// Helper function to safely extract nested string from map
func getNestedString(data map[string]interface{}, keys ...string) string {
	current := data
	for i, key := range keys {
		if val, ok := current[key]; ok {
			if i == len(keys)-1 {
				// Last key, expect string
				return getString(val)
			}
			// Not last key, expect map
			if nextMap, ok := val.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return ""
			}
		} else {
			return ""
		}
	}
	return ""
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

	// Log the incoming data for debugging
	log.Printf("Incoming request data: %+v", requestData)

	// Parse address lines
	addrLine1 := strings.Replace(getNestedString(requestData, "q4_address4", "addr_line1"), "+", " ", -1)
	addrLine2 := strings.Replace(getNestedString(requestData, "q4_address4", "addr_line2"), "+", " ", -1)

	// Combine address lines
	fullAddress := addrLine1
	if addrLine2 != "" {
		if addrLine1 != "" {
			fullAddress += ", " + addrLine2
		} else {
			fullAddress = addrLine2
		}
	}

	// Transform incoming request data to ShelterManager format with safe extraction
	personData := map[string]string{
		"ownertype":        "1",
		"forenames":        getNestedString(requestData, "q3_fullName3", "first"),
		"surname":          getNestedString(requestData, "q3_fullName3", "last"),
		"address":          fullAddress,
		"town":             getNestedString(requestData, "q4_address4", "city"),
		"county":           getNestedString(requestData, "q4_address4", "state"),
		"postcode":         getNestedString(requestData, "q4_address4", "postal"),
		"country":          "USA",
		"hometelephone":    getNestedString(requestData, "q88_mobilePhone", "full"), // Fixed field name
		"emailaddress":     getString(requestData["q6_email6"]),
		"jurisdiction":     "1",
		"flags":            "adopter",
		"gdprcontactoptin": "",
		"site":             "0",
		"a.1.7":            "No",
	}

	// Handle date of birth - fixed field name from q35_dob to q35_dateOf
	month := getNestedString(requestData, "q35_dateOf", "month")
	day := getNestedString(requestData, "q35_dateOf", "day")
	year := getNestedString(requestData, "q35_dateOf", "year")
	if month != "" && day != "" && year != "" {
		personData["dateofbirth"] = fmt.Sprintf("%s/%s/%s", month, day, year)
	}

	// Map the ID from q86_typeA86 field
	if id := getString(requestData["q86_typeA86"]); id != "" {
		personData["idnumber"] = id
	}

	// If mobile phone is empty, try the secondary phone number field
	if personData["hometelephone"] == "" {
		personData["hometelephone"] = getNestedString(requestData, "q116_phoneNumber116", "full")
	}

	// Log the transformed data
	log.Printf("Transformed person data: %+v", personData)

	newPersonURL := baseURL + "/person_new"
	if err := createPerson(newPersonURL, sessionID, personData); err != nil {
		http.Error(w, fmt.Sprintf("Error creating person: %s", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Person created successfully!")
}
