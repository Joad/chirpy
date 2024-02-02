package database

import (
	"os"
	"testing"
)

func TestCreateDB(t *testing.T) {
	filename := "database.json"
	NewDB(filename)
	defer os.Remove(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}
}

func TestCreateChirp(t *testing.T) {
	filename := "database.json"
	db, err := NewDB(filename)
	defer os.Remove(filename)
	if err != nil {
		t.Fatal(err)
	}
	chirp, err := db.CreateChirp("Test", 1)
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
	defer os.Remove(filename)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CreateChirp("Test1", 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CreateChirp("Test2", 1)
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
