package main

import (
	"encoding/json"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"log"
	"okea/core/db"
	"okea/services/models"
	"os"
)

var db_dir_name string
var showusers bool
var uuid string
var showprojects bool
var userfile string
var rmuser string
var rmuuid string
var projectfile string
var rmproject string

func start() (*db.DBHandler, error) {
	log.Printf(" Running okeadb ----\n")
	return db.OpenDBs(db_dir_name)
}

func main() {

	app := cli.NewApp()
	app.Usage = "Manipulate Okea DB files"

	app.Version = "AKASTE 0.0.2"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "databasedir,db",
			Usage:       "Database directory for holding okea information",
			Destination: &db_dir_name,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "status",
			Usage:       "okea status",
			Description: "show status information about okea",
			Action: func(c *cli.Context) {
				println("status ", c.Args().First())
			},
		},
		{
			Name:        "check",
			Usage:       "check up on various aspects of okea service",
			Description: "determine if data and settings are sound",
			Action: func(c *cli.Context) {
				println("check ", c.Args().First())
			},
		},
		{
			Name:  "ls",
			Usage: "Show the various types of info",
			Action: func(c *cli.Context) {
				mdb, _ := start()
				defer mdb.CloseDBs()
				if len(uuid) > 0 {
					user, _ := mdb.GetUserByUUID(uuid)
					if len(user.Username) > 0 {
						log.Printf(" %v/%v/%v\n", user.Username, user.UUID, user.Email)
					} else {
						log.Printf(" No user found\n")
					}
				} else if showusers {
					users, _ := mdb.GetAllUsers()
					for _, user := range users {
						log.Printf(" %v/%v/%v/%v\n", user.Username, user.UUID, user.Email, user.CanAdmin)
					}
				}

			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "users,u",
					Usage:       "Show list of users",
					Destination: &showusers,
				},
				cli.StringFlag{
					Name:        "uuid",
					Usage:       "Show list of users",
					Destination: &uuid,
				},
				cli.BoolFlag{
					Name:        "projects,p",
					Usage:       "Show list of projects",
					Destination: &showprojects,
				},
			},
		},
		{
			Name:  "add",
			Usage: "Add a project or user",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "user,u",
					Usage:       "Add a user, pointing to JSON file that describes the user",
					Destination: &userfile,
				},
				cli.StringFlag{
					Name:        "project,p",
					Usage:       "Add a project, pointing to JSON file that describes the project",
					Destination: &projectfile,
				},
			},
			Action: func(c *cli.Context) {
				mdb, _ := start()
				defer mdb.CloseDBs()
				if len(userfile) > 0 {
					log.Printf("   Loading user data from '%s'\n", userfile)
					js, err := ioutil.ReadFile(userfile)
					if err != nil {
						panic(err)
					}
					users := make([]models.User, 0)
					err = json.Unmarshal(js, &users)
					if err != nil {
						panic(err)
					}
					for _, user := range users {
						log.Printf("    Found user: %s\n", user.Username)
						mdb.AddUser(&user)
					}

					//					if err = mdb.UpdateUser(&user); err != nil {
					//						panic(err)
					//					}
				}
				if len(projectfile) > 0 {
				}
			},
		},
		{
			Name: "rm",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "user,u",
					Usage:       "Remove a user",
					Destination: &rmuser,
					Value:       "",
				},
				cli.StringFlag{
					Name:        "uuid",
					Usage:       "Remove a user by UUID",
					Destination: &rmuuid,
					Value:       "",
				},
				cli.StringFlag{
					Name:        "project,p",
					Usage:       "Remove a project",
					Destination: &rmproject,
					Value:       "",
				},
			},
			Action: func(c *cli.Context) {
				mdb, _ := start()
				defer mdb.CloseDBs()
				if len(rmuser) > 0 {
					user, _ := mdb.GetUserByUsername(rmuser)
					if user == nil {
						log.Printf(" Unable to locate user: %v\n", rmuser)
					} else {
						log.Printf(" Found user and deleting: %v/%v\n", user.Username, user.UUID)
						err := mdb.DeleteUserByUUID(user.UUID)
						if err != nil {
							log.Printf("  error: %v\n", err)
						}
					}
				} else if len(rmuuid) > 0 {
					err := mdb.DeleteUserByUUID(rmuuid)
					if err != nil {
						log.Printf("  error: %v\n", err)
					}

				} else if len(rmproject) > 0 {
				}
			},
		},
	}

	// This works effectively as the main of the program
	app.Action = func(c *cli.Context) {
	}

	app.Run(os.Args)
}
