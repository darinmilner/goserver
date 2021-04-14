package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/darinmilner/goserver/internal/models"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name   string
	url    string
	method string

	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	// {"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	// {"post-search-avail", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2021-01-01"},
	// 	{key: "end", value: "2021-01-02"},
	// }, http.StatusOK},
	// {"post-search-avail-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2021-01-01"},
	// 	{key: "end", value: "2021-01-02"},
	// }, http.StatusOK},
	// {"make-reservation-post", "/make-reservation", "POST", []postData{
	// 	{key: "first-name", value: "Abdillah"},
	// 	{key: "last-name", value: "Ali"},
	// 	{key: "email", value: "aa@amail.com"},
	// 	{key: "phone", value: "123-333-456"},
	// }, http.StatusOK},
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusOK)
	}

	//test case with non existant room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()

	reservation.RoomID = 1000
	session.Put(ctx, "reservation", reservation)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusOK)
	}

}

func TestRpositoryChooseRoom(t *testing.T) {
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

	if rr.Code != http.StatusTemporaryRedirect {
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

	if rr.Code != http.StatusTemporaryRedirect {
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

	if rr.Code != http.StatusTemporaryRedirect {
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned response code got %d but wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned response code for invalid startdate: %d but wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned response code for invalid enddate: %d but wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned response code for invalid roomId: %d but wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned response code for invalid data: %d but wanted %d", rr.Code, http.StatusSeeOther)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed inserting data into db: %d but wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned response code for invalid enddate: %d but wanted %d", rr.Code, http.StatusTemporaryRedirect)
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

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}
	return ctx
}
