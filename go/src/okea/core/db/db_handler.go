package db

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"okea/services/models"
	"okea/settings"
	"path/filepath"
	"strings"
)

const UserDBFileName = "users.bolt"
const ProjectDBFileName = "projects.bolt"

const ProjectBucket = "projects"
const UserBucket = "users"

type DBHandler struct {
	Directory string
	userDB    *bolt.DB
	projectDB *bolt.DB
}

func (db *DBHandler) OpenUserDB() error {
	var err error
	userdbpath := filepath.Join(db.Directory, UserDBFileName)
	if db.userDB, err = bolt.Open(userdbpath, 0600, nil); err != nil {
		log.Fatalf(" Unable to open user DB '%s': ", userdbpath)
		log.Fatalf("'%v'\n", err)
		return err
	}
	if settings.VerboseLog {
		log.Printf("  opened user DB '%s'\n", userdbpath)
	}

	db.userDB.Update(func(tx *bolt.Tx) error {
		var err error
		if _, err = tx.CreateBucketIfNotExists([]byte(UserBucket)); err != nil {
			log.Fatal(err)
			return err
		}
		if settings.VerboseLog {
			log.Printf("   found '%s' bucket\n", UserBucket)
		}
		return nil
	})

	return nil
}

func (db *DBHandler) OpenProjectDB() error {
	var err error
	projectdbpath := filepath.Join(db.Directory, ProjectDBFileName)
	if db.projectDB, err = bolt.Open(projectdbpath, 0600, nil); err != nil {
		log.Fatalf(" Unable to open project DB '%s': %v\n", projectdbpath, err)
		return err
	}
	if settings.VerboseLog {
		log.Printf("  opened project DB '%s'\n", projectdbpath)
	}

	db.projectDB.Update(func(tx *bolt.Tx) error {
		var err error
		if _, err = tx.CreateBucketIfNotExists([]byte(ProjectBucket)); err != nil {
			log.Fatal(err)
			return err
		}
		if settings.VerboseLog {
			log.Printf("   found '%s' bucket\n", ProjectBucket)
		}
		return nil
	})

	return nil
}

func OpenDBs(db_dir string) (*DBHandler, error) {
	d := new(DBHandler)
	d.Directory = db_dir
	d.OpenUserDB()
	d.OpenProjectDB()

	if settings.VerboseLog {
		s := d.StatsAsJSON()
		log.Printf("  db stats: %s\n", s)
	}

	return d, nil
}

func (db *DBHandler) CloseDBs() error {
	if err := db.userDB.Close(); err != nil {
		log.Fatal(err)
		return err
	}
	if err := db.projectDB.Close(); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (db *DBHandler) StatsAsJSON() string {
	us := db.userDB.Stats()
	ps := db.projectDB.Stats()
	ju, _ := json.Marshal(us)
	jp, _ := json.Marshal(ps)
	return "{\"user\":" + string(ju) + "},{\"project\":" + string(jp) + "}"
}

func (db *DBHandler) UpdateUser(user *models.User) error {

	var buf bytes.Buffer
	gob.Register(models.User{})

	return db.userDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucket))
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(user); err != nil {
			panic(err)
		}
		log.Printf(" uuid: %s\n", user.UUID)
		err := b.Put([]byte(user.UUID), buf.Bytes())
		return err
	})
}

func (db *DBHandler) AddUser(user *models.User) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	user.Password = string(hashed)
	user.UUID = uuid.New()
	db.UpdateUser(user)
	return nil
}

func (db *DBHandler) UpdateProject(project *models.Project) error {
	return db.projectDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ProjectBucket))
		projectbuf := &bytes.Buffer{}
		if err := binary.Write(bufio.NewWriter(projectbuf), binary.BigEndian, project); err != nil {
			panic(err)
		}
		return b.Put([]byte(project.UUID), projectbuf.Bytes())
	})
}

func (db *DBHandler) GetUserByUsername(username string) (*models.User, error) {
	founduser := new(models.User)
	users, err := db.GetAllUsers()
	if err != nil {
		panic(err)
	}
	for _, user := range users {
		if strings.Compare(username, user.Username) == 0 {
			*founduser = user
			return founduser, nil
		}
	}

	return nil, errors.New("User not found")
}

func (db *DBHandler) DeleteUserByUUID(uuid string) error {
	dberr := db.userDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucket))
		return b.Delete([]byte(uuid))
	})

	return dberr
}

func (db *DBHandler) GetUserByUUID(uuid string) (*models.User, error) {
	user := new(models.User)
	dberr := db.userDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucket))
		gob.Register(models.User{})
		v := b.Get([]byte(uuid))
		dec := gob.NewDecoder(bytes.NewBuffer(v))
		dec.Decode(user)
		return nil
	})

	return user, dberr
}

func (db *DBHandler) GetAllUsers() ([]models.User, error) {
	var users []models.User
	dberr := db.userDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucket))
		gob.Register(models.User{})
		b.ForEach(func(k, v []byte) error {
			newone := &models.User{}
			dec := gob.NewDecoder(bytes.NewBuffer(v))
			dec.Decode(&newone)
			users = append(users, *newone)
			return nil
		})
		return nil
	})
	return users, dberr
}
