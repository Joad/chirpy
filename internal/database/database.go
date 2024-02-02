package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

type Chirp struct {
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
	Body     string `json:"body"`
}

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps      map[int]Chirp        `json:"chirps"`
	Users       map[int]User         `json:"users"`
	Revocations map[string]time.Time `json:"revocations"`
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
			Chirps:      make(map[int]Chirp),
			Users:       make(map[int]User),
			Revocations: make(map[string]time.Time),
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

func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := len(dbstruct.Chirps) + 1

	newChirp := Chirp{
		Id:       id,
		AuthorId: authorId,
		Body:     body,
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

func (db *DB) DeleteChirp(id int) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return err
	}
	delete(dbstruct.Chirps, id)
	db.writeDB(dbstruct)
	return nil
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

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, ok := dbstruct.Users[id]
	if !ok {
		return User{}, errors.New("User not found")
	}
	user.Email = email
	user.Password = password
	dbstruct.Users[id] = user
	err = db.writeDB(dbstruct)
	return user, err
}

func (db *DB) UpgradeUser(id int) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return err
	}
	user, ok := dbstruct.Users[id]
	if !ok {
		return errors.New("User not found")
	}
	user.IsChirpyRed = true
	dbstruct.Users[id] = user
	return db.writeDB(dbstruct)
}

func (db *DB) IsTokenRevoked(token string) (bool, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return false, err
	}
	_, found := dbstruct.Revocations[token]
	return found, nil
}

func (db *DB) RevokeToken(token string, revocationTime time.Time) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbstruct, err := db.loadDB()
	if err != nil {
		return err
	}

	dbstruct.Revocations[token] = revocationTime
	return db.writeDB(dbstruct)
}
