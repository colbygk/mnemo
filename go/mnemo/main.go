package main

import (
	//	"bytes"
	//	"crypto/tls"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
	//"io/ioutil"
	"log"
	"mnemo/api"
	//	"net/http"
	"github.com/bgentry/speakeasy"
	"okea/core/config"
	"os"
	"os/user"
)

var Verbose bool
var config_filename string

var okea_hostname string
var okea_port int
var okea_username string
var okea_password string
var okea_trusttls bool

func get_local_credentials(c *cli.Context) (api.Credentials, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return api.Credentials{}, err
	}

	return api.FileToToken(fmt.Sprintf("%s/.mnemo/token.json", usr.HomeDir))
}

func init_local_credentials(c *cli.Context) (*os.File, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = os.MkdirAll(fmt.Sprintf("%s/.mnemo", usr.HomeDir), 0700)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return os.OpenFile(fmt.Sprintf("%s/.mnemo/token.json", usr.HomeDir), os.O_RDWR|os.O_CREATE, 0600)
}

func init_project(app *cli.App, c *cli.Context) {
	conf := init_cli(app, c)
	cred, err := get_local_credentials(c)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to read local credentials: %v\n", err))
	}

	println("init project placeholder", c.Args().First())

	oks := &api.OkeaService{
		Hostname: conf.Hostname,
		Port:     conf.Port,
		Username: conf.Okea_username,
		Password: "testing",
		TrustTLS: conf.TrustTLS,
		Token:    cred.Token}

	resp, herr := oks.Hello()
	if herr != nil {
		log.Fatal(fmt.Sprintf("Unable to read local credentials: %v\n", herr))
	}
	log.Printf("oh hai: %v\n", resp)

	resp, herr = oks.ListProjects()
	if herr != nil {
		log.Fatal(fmt.Sprintf("Unable to read local credentials: %v\n", herr))
	}
	log.Printf("oh hai: %v\n", resp)
}

func init_cli(app *cli.App, c *cli.Context) config.JSONConfig {
	if Verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		o, err := os.OpenFile(string(app.Name+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf(string("Unable to open file: " + app.Name + ".log"))
		}
		log.SetOutput(o)
		log.Println("------- Starting " + app.Name + " -------")
	}
	conf, err := config.LoadConfig(config_filename)

	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		jconf := config.JSONConfig{
			Hostname:      okea_hostname,
			Port:          okea_port,
			Okea_username: okea_username,
			TrustTLS:      okea_trusttls}

		usr, uerr := user.Current()
		if uerr != nil {
			log.Fatal(uerr)
		}

		err = os.MkdirAll(fmt.Sprintf("%s/.mnemo", usr.HomeDir), 0700)
		if err != nil {
			log.Fatal(err)
		}

		err = jconf.WriteConfig(config_filename)
		if err != nil {
			log.Fatalf("Unable to save default config: %v\n", err)
		}

		conf = jconf

	} else if err != nil {
		log.Fatal(fmt.Sprintf("Error loading config: %v\n", err))
	}
	if Verbose {
		log.Printf("remote API https://%v:%v\n", conf.Hostname, conf.Port)
		log.Printf("remote API ignore certificate problems: %v\n", conf.TrustTLS)
	}

	return conf
}

func main() {
	var lang_type string
	var frontend string

	app := cli.NewApp()
	app.Usage = "Tool to create, run and deploy projects."
	// Nymphs associated with Nephelei nymphs:
	// The eldest Okeanides: numbered among the Titanides
	// - Dione Doris Elektra Eurynome Klymene Metis Neda Pleione Styx.
	// These were most likely heavenly goddesses of the clouds.
	// Other names:
	// ADMETE, AKASTE, AMPHIRO, AMPHITRITE, ARGIA, ASIA, BEROE,
	// DIONE, DORIS, EIDYIA, ELEKTRA, EPHYRA, EUAGOREIS, EUDORA,
	// EUROPA, EURYNOME, GALAXAURA GALAXAURA, HIPPO, IAKHE, IANEIRA,
	// IANTHE, KALLIRHOE, KALYPSO, KERKEIS, KHRYSEIS, KLEIO, KLYMENE,
	// KLYTIA, LEUKIPPE, MELIA, MELIBOEA, MELITE, MELOBOSIS, M
	// ENIPPE, METIS, OKYRHOE, OKYROE, OURANIA, PASIPHAE, PASITHOE,
	// PEITHO, PERSEIS, PETRIAE, PHAINO, PLEIO NE PLEIONE, PLEXAURE,
	// PLOUTO, POLYDORA, POLYXO, PRYMNO, RHODEA, RHODEIA, RHODOPE,
	// STILBO, STYX, TELE STO, THOE, TYKHE, XANTHE, ZEUXO
	// Source: http://www.theoi.com/Nymphe/Okeanides.html

	app.Version = "AKASTE 0.0.2"

	usr, uerr := user.Current()
	if uerr != nil {
		log.Fatal(uerr)
	}
	okea_username = usr.Username
	mnemo_conf_name := fmt.Sprintf("%s/.mnemo/config.json", usr.HomeDir)

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "Verbose, V",
			Usage:       string("Log verbose output to " + app.Name + ".log"),
			Destination: &Verbose,
		},
		cli.StringFlag{
			Name:        "config,c",
			Value:       mnemo_conf_name,
			Usage:       "Read config from file, JSON",
			Destination: &config_filename,
			EnvVar:      "NEPH_CONFIG",
		},
		cli.StringFlag{
			Name:        "host",
			Value:       "localhost",
			Usage:       "Okea service to communicate with",
			Destination: &okea_hostname,
			EnvVar:      "OKEA_HOSTNAME",
		},
		cli.IntFlag{
			Name:        "port,p",
			Value:       12098,
			Usage:       "Okea service to communicate with",
			Destination: &okea_port,
			EnvVar:      "OKEA_PORT",
		},
		cli.StringFlag{
			Name:        "username,u",
			Value:       okea_username,
			Usage:       "Okea username login",
			Destination: &okea_username,
			EnvVar:      "OKEA_USERNAME",
		},
		cli.BoolFlag{
			Name:        "trusttls,tls",
			Usage:       "Okea trust certificate blindly, DANGER",
			Destination: &okea_trusttls,
			EnvVar:      "OKEA_TRUSTTLS",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "deploy",
			Aliases:     []string{"d"},
			Usage:       "mnemo deploy blah blah blah",
			Description: "deploy this project",
			Action: func(c *cli.Context) {
				init_cli(app, c)
				println("deploy ", c.Args().First())
			},
		},
		{
			Name:        "init",
			Aliases:     []string{"i"},
			Usage:       "mnemo init blah blah blah",
			Description: "initialize a new project",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "type,t",
					Value:       "node",
					Usage:       "code base of project, one of go|swift|node|ruby|python",
					Destination: &lang_type,
					EnvVar:      "NEPH_PROJ_TYPE",
				},
				cli.StringFlag{
					Name:        "frontend,f",
					Value:       "ml-web",
					Usage:       "Frontend, one of ml-web|byo-web|direct",
					Destination: &frontend,
					EnvVar:      "NEPH_FRONTEND",
				},
			},
			Action: func(c *cli.Context) {
				init_project(app, c)
			},
			// Subcommands: {},
		},
		{
			Name:        "login",
			Aliases:     []string{"l"},
			Usage:       "mnemo login",
			Description: "Log into Nephelei server",
			Action: func(c *cli.Context) {

				okea_password, perr := speakeasy.Ask("password: ")
				if perr != nil {
					log.Fatal(perr)
				}

				conf := init_cli(app, c)
				creds_file, cerr := init_local_credentials(c)

				if cerr != nil {
					log.Fatal(cerr)
				}
				log.Printf(" init_cli: %v\n", conf)
				oks := &api.OkeaService{
					Hostname: conf.Hostname,
					Port:     conf.Port,
					Username: conf.Okea_username,
					Password: okea_password,
					TrustTLS: conf.TrustTLS,
					Token:    ""}

				usr, err := user.Current()
				if err != nil {
					log.Fatal(err)
				}
				err = os.MkdirAll(fmt.Sprintf("%s/.mnemo", usr.HomeDir), 0700)
				if err != nil {
					log.Fatal(err)
				}

				conf.WriteConfig(fmt.Sprintf("%s/.mnemo/config.json", usr.HomeDir))
				resp, rerr := oks.Login()
				if Verbose {
					log.Printf("login resp/err; %v/%v", resp, rerr)
				}

				if rerr == nil {
					creds_file.WriteString(resp)
					creds_file.Sync()
				}
			},
		},
		{
			Name:        "ls",
			Usage:       "mnemo ls",
			Description: "List your projects",
			Action: func(c *cli.Context) {
				//projectlist, err := okea.ls()
			},
		},
		{
			Name:        "run",
			Aliases:     []string{"r"},
			Usage:       "mnemo run blah blah",
			Description: "run your superawesomeness",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name,n",
					Value: "generandomnamehere",
					Usage: "random nameything",
				},
			},
			Action: func(c *cli.Context) {
				println("run ", c.Args().First())
			},
		},
	}

	app.Run(os.Args)
}
