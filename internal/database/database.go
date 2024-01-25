package database

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}

	err := db.ensureDB()
	return db, err
}

func (db *DB) ensureDB() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		return db.writeDB(DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		})
	} else if err != nil {
		return err
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	dat, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	structure := DBStructure{}
	err = json.Unmarshal(dat, &structure)
	if err != nil {
		return DBStructure{}, err
	}
	return structure, nil
}

func (db *DB) writeDB(dbstruct DBStructure) error {
	dat, err := json.Marshal(dbstruct)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, dat, 0600)
	return err
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := len(dbstruct.Chirps) + 1

	newChirp := Chirp{
		Id:   id,
		Body: body,
	}
	dbstruct.Chirps[id] = newChirp

	err = db.writeDB(dbstruct)
	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := make([]Chirp, 0, len(dbstruct.Chirps))
	for _, chirp := range dbstruct.Chirps {
		chirps = append(chirps, chirp)
	}
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})
	return chirps, nil
}

func (db *DB) GetChirpById(id int) (Chirp, bool, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, false, err
	}

	chirp, found := dbstruct.Chirps[id]
	if !found {
		return Chirp{}, false, nil
	}
	return chirp, true, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if _, found := findUserByEmail(dbstruct.Users, email); found {
		return User{}, errors.New("User already exists with email")
	}

	id := len(dbstruct.Users) + 1

	newUser := User{
		Id:       id,
		Email:    email,
		Password: password,
	}

	dbstruct.Users[id] = newUser

	err = db.writeDB(dbstruct)
	return newUser, err
}

func findUserByEmail(users map[int]User, email string) (User, bool) {
	for _, user := range users {
		if user.Email == email {
			return user, true
		}
	}
	return User{}, false
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, found := findUserByEmail(dbstruct.Users, email)
	if !found {
		return User{}, errors.New("User not found")
	}

	return user, nil
}
