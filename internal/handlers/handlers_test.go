package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/darinmilner/goserver/internal/driver"
	"github.com/darinmilner/goserver/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"non-existent", "/green/eggs/notfound", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"new reservation", "/admin/new-reservations", "GET", http.StatusOK},
	{"all reservation", "/admin/all-reservations", "GET", http.StatusOK},
	{"show one reservation", "/admin/reservations/new/1/show", "GET", http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)

	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}

		}
	}
}

// testAvailabilityJSONData is data for the AvailabilityJSON handler, /search-availability-json route
var testAvailabilityJSONData = []struct {
	name            string
	postedData      url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start":   {"2050-01-01"},
			"end":     {"2050-01-02"},
			"room_id": {"1"},
		},
		expectedOK: false,
	}, {
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedOK: true,
	},
	{
		name:            "empty post body",
		postedData:      nil,
		expectedOK:      false,
		expectedMessage: "Internal Server Error",
	},
	{
		name: "database query fails",
		postedData: url.Values{
			"start":   {"2060-01-01"},
			"end":     {"2060-01-02"},
			"room_id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Error querying database",
	},
}

// TestAvailabilityJSON tests the AvailabilityJSON handler
func TestAvailabilityJSON(t *testing.T) {
	for _, e := range testAvailabilityJSONData {
		// create request, get the context with session, set header, create recorder
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// make our handler a http.HandlerFunc and call
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(rr, req)

		var j jsonResponse
		err := json.Unmarshal([]byte(rr.Body.String()), &j)
		if err != nil {
			t.Error("failed to parse json!")
		}

		if j.OK != e.expectedOK {
			t.Errorf("%s: expected %v but got %v", e.name, e.expectedOK, j.OK)
		}
	}
}

func TestRepositoryReservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)

	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusOK)
	}

	//test case where reservation is not in session (reset)
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)

	//session without reservation
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test case with non existant room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	reservation.RoomID = 1000
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusOK)
	}

}

func TestRepositoryChooseRoom(t *testing.T) {
	//Reservation in Context
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}
	//Get Context
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	//Make test recorder
	rr := httptest.NewRecorder()

	//Put reservation in context
	session.Put(ctx, "reservation", reservation)

	//Server HTTP handlefunc
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Wrong response code returned got %d", rr.Code)
	}

	//NO Reservation in Context
	//Get Context
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"

	//Make test recorder
	rr = httptest.NewRecorder()

	//No Session

	//Server HTTP handlefunc
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Wrong response code returned got %d", rr.Code)
	}

	//Incorrect Room Number
	//Get Context
	req, _ = http.NewRequest("GET", "/choose-room/room", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/room"

	//Make test recorder
	rr = httptest.NewRecorder()

	//No Session

	//Server HTTP handlefunc
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Wrong response code returned got %d", rr.Code)
	}
}

func TestRepositoryReservationSummary(t *testing.T) {
	//Reservation in Context
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	//Get Context
	req, _ := http.NewRequest("POST", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	//Make test recorder
	rr := httptest.NewRecorder()

	//Put reservation in context
	session.Put(ctx, "reservation", reservation)

	//Server HTTP handlefunc
	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response code returned got %d", rr.Code)
	}

	//No Reservation in Session
	req, _ = http.NewRequest("POST", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	//Make test recorder
	rr = httptest.NewRecorder()

	//Put reservation in context
	//session.Put(ctx, "reservation", reservation)

	//Server HTTP handlefunc
	handler = http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Wrong response code returned got %d", rr.Code)
	}

}

func TestRepositoryPostReservation(t *testing.T) {

	reqBody := "start-date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=Ali")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	postData := url.Values{}
	postData.Add("start-date", "2050-01-01")
	postData.Add("end-date", "2050-01-03")
	postData.Add("first-name", "Yusuf")
	postData.Add("last-name", "Grenada")
	postData.Add("email", "yg@yg.com")
	postData.Add("phone", "222-122-0122")
	postData.Add("room-id", "1")

	//postData.Encode()
	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//Test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//Test for invalid start date
	reqBody = "start-date=invalid"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=Ali")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned response code for invalid startdate: %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for invalid end date
	reqBody = "start-date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=invalid")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=Ali")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned response code for invalid enddate: %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for invalid room Id
	reqBody = "start-date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=Ali")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=invalid")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned response code for invalid roomId: %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for invalid data
	reqBody = "start-date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=j")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("PostReservation handler returned response code for invalid data: %d but wanted %d", rr.Code, http.StatusOK)
	}

	//test for failure to insert data to the db
	reqBody = "start-date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=Ali")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=2")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler failed inserting data into db: %d but wanted %d", rr.Code, http.StatusSeeOther)
	}

	//test for failure to insert restriction into db
	reqBody = "start-date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end-date=invalid")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first-name=Ali")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last-name=Jamal")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=aJamal@abc.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=200000")

	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx = getCtx(req)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()

	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned response code for invalid enddate: %d but wanted %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepositoryAvailabilityJSON(t *testing.T) {

	//Rooms are not available
	reqBody := "start=2070-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2070-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	//create request

	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	//get contest with session
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	//set header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	//make handler func

	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	//Make request to our handler
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var j jsonResponse
	err := json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("Failed to parse JSON")
	}

	if j.OK {
		t.Error("Got availability when none was expected")
	}

	//Rooms are available
	reqBody = "start=2040-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2040-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	//create request

	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	//get contest with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	//set header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	//make handler func

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	//Make request to our handler
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	//parse JSON to get expected response
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("Failed to parse JSON")
	}
	if !j.OK {
		t.Error("Got no availability when availability was expected")
	}

	//NIL request
	req, _ = http.NewRequest("POST", "/search-availability-json", nil)

	//get contest with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	//set header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	//make handler func
	//Make request to our handler
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	handler.ServeHTTP(rr, req)
	log.Print(&j)

	//parse JSON to get expected response
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("Failed to Parse JSON")
	}

	if j.OK || j.Message != "Internal Server Error" {
		t.Error("No request body returns ok json")
	}

	//Rooms are NOT available
	reqBody = "start=2080-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2080-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room-id=1")

	//create request

	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	//get contest with session
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	//set header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	//make handler func

	handler = http.HandlerFunc(Repo.AvailabilityJSON)

	//Make request to our handler
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	//parse JSON to get expected response
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Error("Failed to parse JSON")
	}
	if j.OK {
		t.Error("Got availability when no availability was expected")
	}
}

var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{"valid-credentials",
		"me@me.com",
		http.StatusSeeOther,
		"",
		"/"},
	{"Invalid-credentials",
		"abdillah@me.com",
		http.StatusSeeOther,
		"",
		"/user/login"},
	{"Invalid-data",
		"a",
		http.StatusOK,
		`action="/user/login`,
		""},
}

func TestLogin(t *testing.T) {
	//range through all test
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		//set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		//call the handler
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s expected code %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s but got location %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		if e.expectedHTML != "" {
			//read the response body to a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s, expected to find %s", e.name, e.expectedHTML)
			}
		}

	}
}

// testPostAvailabilityData is data for the PostAvailability handler test, /search-availability
var testPostAvailabilityData = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "rooms not available",
		postedData: url.Values{
			"start": {"2050-01-01"},
			"end":   {"2050-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "rooms are available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "invalid start data",
		postedData: url.Values{
			"start": {"2022BB-01-01"},
			"end":   {"2022-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "empty data",
		postedData: url.Values{
			"start": {""},
			"end":   {"2022-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
}

func TestPostRoomAvailability(t *testing.T) {
	for _, e := range testPostAvailabilityData {
		req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(e.postedData.Encode()))

		//Get Context
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		//set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		//call the handler
		handler := http.HandlerFunc(Repo.PostAvailability)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s gave wrong status code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

// bookRoomTests is the data for the BookRoom handler tests
var bookRoomTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{

		name:               "database-works",
		url:                "/book-room?s=2050-01-01&e=2050-01-02&id=1",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "database-fails",
		url:                "/book-room?s=2040-01-01&e=2040-01-02&id=2",
		expectedStatusCode: http.StatusSeeOther,
	},
}

//TestBookRoom tests the BookRoom handler
func TestBookRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	for _, e := range bookRoomTests {
		req, _ := http.NewRequest("GET", e.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		session.Put(ctx, "reservation", reservation)

		handler := http.HandlerFunc(Repo.BookRoom)

		log.Print("Before handle func ", rr, req)
		handler.ServeHTTP(rr, req)
		log.Print("After handle func ", rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("%s failed: returned wrong response code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

//reservationSummaryTests
var reseversationSummaryTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "res-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/search-availability",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/reservation-summary",
	},
	{
		name:               "res-not-in-session",
		reservation:        models.Reservation{},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

func TestReservationSummary(t *testing.T) {
	for _, e := range reseversationSummaryTests {
		req, _ := http.NewRequest("GET", e.url, nil)

		//Get Context
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		//set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		//call the handler
		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(rr, req)

		log.Println(rr)
		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s gave wrong status code: got %d, wanted %d", e.name, rr.Code, e.expectedStatusCode)
		}
	}
}

var adminPostReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "cal-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "cal-res",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

func TestPostReservationCalendar(t *testing.T) {
	for _, e := range adminPostReservationCalendarTests {
		var req *http.Request
		if e.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/calendar", strings.NewReader(e.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/calendar", nil)
		}
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currentYear, currentMonth, _ := now.Date()
		currentLocation := now.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if e.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = e.blocks
		}

		if e.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = e.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

	}
}

var adminShowReservationsTests = []struct {
	name                 string
	expectedResponseCode int
	expectedHTML         string
}{
	{
		name:                 "Reservations list",
		expectedResponseCode: http.StatusOK,
		expectedHTML:         "<th>Last Name</th>",
	},
}

func TestAdminShowReservations(t *testing.T) {
	for _, e := range adminShowReservationsTests {
		var req *http.Request

		req, _ = http.NewRequest("GET", "/all-reservations", nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminAllReservations)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}

	}
}

func TestAdminShowNewReservations(t *testing.T) {
	for _, e := range adminShowReservationsTests {
		var req *http.Request

		req, _ = http.NewRequest("GET", "/new-reservations", nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		// set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// call the handler
		handler := http.HandlerFunc(Repo.AdminNewReservations)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedResponseCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}

		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}

	}
}

var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "process-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "process-reservation-back-to-cal",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

func TestAdminProcessReservation(t *testing.T) {
	for _, e := range adminProcessReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", e.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedResponseCode, rr.Code)
		}
	}
}

func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *Repository", reflect.TypeOf(testRepo).String())
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}
	return ctx
}
