package dbrepo

import (
	"errors"
	"log"
	"time"

	"github.com/darinmilner/goserver/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

//InsertReservation inserts a reservation to the DB
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {

	if res.RoomID == 2 {
		return 0, errors.New("An error occurred")
	}
	return 1, nil
}

//InsertRoomRestriction inserts a room restriction into the DB
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {

	if r.RoomID == 200_000 {
		return errors.New("An error occurred")
	}
	return nil
}

//SearchAvailabilityByRoomID returns true if roomID room is available and false if not available
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {

	layout := "2006-01-02"
	str := "2049-12-31"
	t, err := time.Parse(layout, str)
	if err != nil {
		log.Println(err)
	}

	// this is our test to fail the query -- specify 2060-01-01 as start
	testDateToFail, err := time.Parse(layout, "2060-01-01")
	if err != nil {
		log.Println(err)
	}

	if start == testDateToFail {
		return false, errors.New("TestDate is not available")
	}

	// if the start date is after 2049-12-31, then return false,
	// indicating no availability;
	if start.After(t) {
		return false, nil
	}

	// otherwise, we have availability
	return true, nil
}

//SearchAvailabilityForAllRooms return a slice of available rooms on a date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	var rooms []models.Room

	return rooms, nil
}

//GetRoomByID gets a room by ID
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("An error")
	}

	return room, nil

}

func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User

	return u, nil
}

func (m *testDBRepo) UpdateUser(models.User) error {

	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 1, "", nil
}
