package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// setupTestData initializes or resets the UserData for a clean test run.
func setupTestData() {
	mutex = &sync.Mutex{}
	UserData = make(map[int64][]string)
}

func TestCountryHandlers(t *testing.T) {
	setupTestData()

	// 1. Test Add Country
	addReqBody := `{"userId": 789, "country": "France"}`
	addReq, _ := http.NewRequest("POST", "/api/countries", bytes.NewBufferString(addReqBody))
	addRr := httptest.NewRecorder()
	http.HandlerFunc(addCountry).ServeHTTP(addRr, addReq)

	if status := addRr.Code; status != http.StatusCreated {
		t.Errorf("addCountry handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if len(UserData[789]) != 1 || UserData[789][0] != "France" {
		t.Errorf("UserData not updated correctly after add: got %v", UserData[789])
	}

	// 2. Test Get Country
	getReq, _ := http.NewRequest("GET", "/api/countries?userId=789", nil)
	getRr := httptest.NewRecorder()
	http.HandlerFunc(getCountries).ServeHTTP(getRr, getReq)

	if status := getRr.Code; status != http.StatusOK {
		t.Errorf("getCountries handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var countries []string
	if err := json.Unmarshal(getRr.Body.Bytes(), &countries); err != nil {
		t.Fatalf("could not parse getCountries response body: %v", err)
	}
	if len(countries) != 1 || countries[0] != "France" {
		t.Errorf("getCountries handler returned unexpected body: got %v want %v", getRr.Body.String(), `["France"]`)
	}

	// 3. Test Delete Country
	delReqBody := `{"userId": 789, "country": "France"}`
	delReq, _ := http.NewRequest("DELETE", "/api/countries", bytes.NewBufferString(delReqBody))
	delRr := httptest.NewRecorder()
	http.HandlerFunc(deleteCountry).ServeHTTP(delRr, delReq)

	if status := delRr.Code; status != http.StatusOK {
		t.Errorf("deleteCountry handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if len(UserData[789]) != 0 {
		t.Errorf("UserData not updated correctly after delete: got %v, want empty", UserData[789])
	}

	// 4. Test Get After Delete
	getAfterDeleteReq, _ := http.NewRequest("GET", "/api/countries?userId=789", nil)
	getAfterDeleteRr := httptest.NewRecorder()
	http.HandlerFunc(getCountries).ServeHTTP(getAfterDeleteRr, getAfterDeleteReq)

	if status := getAfterDeleteRr.Code; status != http.StatusOK {
		t.Errorf("getCountries handler returned wrong status code after delete: got %v want %v", status, http.StatusOK)
	}

	if err := json.Unmarshal(getAfterDeleteRr.Body.Bytes(), &countries); err != nil {
		t.Fatalf("could not parse getCountries response body after delete: %v", err)
	}
	if len(countries) != 0 {
		t.Errorf("getCountries handler returned unexpected body after delete: got %v want %v", getAfterDeleteRr.Body.String(), `[]`)
	}
}
