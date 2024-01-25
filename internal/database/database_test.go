package database

import (
	"os"
	"testing"
)

func TestCreateDB(t *testing.T) {
	filename := "database.json"
	NewDB(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}
	os.Remove(filename)
}

func TestCreateChirp(t *testing.T) {
	filename := "database.json"
	db, err := NewDB(filename)
	if err != nil {
		t.Fatal(err)
	}
	chirp, err := db.CreateChirp("Test")
	if err != nil {
		t.Fatal(err)
	}
	if chirp.Body != "Test" {
		t.Fatalf("Body was %s, expected %s", chirp.Body, "Test")
	}
}

func TestGetChirps(t *testing.T) {
	filename := "database.json"
	db, err := NewDB(filename)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CreateChirp("Test1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CreateChirp("Test2")
	if err != nil {
		t.Fatal(err)
	}

	chirps, err := db.GetChirps()
	if err != nil {
		t.Fatal(err)
	}

	if len(chirps) != 2 {
		t.Fatal("Wrong amount of chirps")
	}
}
