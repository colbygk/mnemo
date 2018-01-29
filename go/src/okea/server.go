package main

/*
 * curl -i -H "Accept: application/json" -H "Content-type: application/json" -X GET https://draco.media.mit.edu:12098/test/hello
  expect a 401

	Get a token:
 * curl -i -H "Content-type: application/json" -X POST -d '{"Username":"haku","Password":"testing"}' https://draco.media.mit.edu:12098/token-auth
  expect a token
	 curl -H "Authorization: Bearer ... token ..." https://draco.media.mit.edu:12098/test/hello

*/

import (
	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/codegangsta/negroni"
	"okea/core"
	//	"okea/core/config"
	"okea/core/db"
	"okea/routers"
	"okea/services"
	"okea/settings"
	//	"github.com/fatih/color"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var nginxdir string
var db_dir_name string
var nginx_yaml_file_name string
var db_handle *bolt.DB
var listen_port int
var listen_ipaddress string
var auth_token string
var config_info string
var docker_endpoint string

type handlerError struct {
	Error   error
	Message string
	Code    int
}

// a custom type that we can use for handling errors and formatting responses
type handler func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError)

type WebReference struct {
	Project string `json:"project"`
	FQDN    string `json:"fqdn"`
	Enabled bool   `json:"enabled"`
	CName   string `json:"cname"`
	ARecord string `json:"arecord"`
	Port    int    `json:"port"`
}

type Status struct {
	status string `json:"status"`
}

/*
nginx_vhosts:
  - listen: "80 default_server"
    server_name: "example.com"
    root: "/var/www/example.com"
    index: "index.php index.html index.htm"
    error_page: ""
    access_log: ""
    error_log: ""
    extra_parameters: |
      location ~ \.php$ {
        fastcgi_split_path_info ^(.+\.php)(/.+)$;
        fastcgi_pass unix:/var/run/php5-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
      }
*/

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// here we could do some prep work before calling the handler if we wanted to

	// call the actual handler
	response, err := fn(w, r)

	// check for errors
	if err != nil {
		log.Printf("ERROR: %v\n", err.Error)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		return
	}
	if response == nil {
		log.Printf("ERROR: response from method is nil\n")
		http.Error(w, "Internal server error. Check the logs.", http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	// send the response and log
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

// NB: Move into separate infoblox API library
func updateCname(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	vars := mux.Vars(r)

	if len(vars["cname"]) == 0 {
		return WebReference{"error:", vars["cname"], false, "http.media.mit.edu", "", 8999}, nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	authinfo := strings.Split(auth_token, ":")
	if settings.VerboseLog {
		log.Printf("len(authinfo): %d", len(authinfo))
		log.Printf("username: %s", authinfo[0])
	}

	var jsonStr string
	jsonStr = fmt.Sprintf("{\"name\":\"%s.media.mit.edu\",\"canonical\":\"http.media.mit.edu\"}", vars["cname"])
	req, _ := http.NewRequest("POST", "https://infoblox2.media.mit.edu/wapi/v1.2/record:cname", bytes.NewBuffer([]byte(jsonStr)))
	req.SetBasicAuth(authinfo[0], authinfo[1])
	req.Header.Set("Content-Type", "application/json")

	resp, rerr := client.Do(req)
	if rerr != nil {
		log.Printf("error updateCname: %s", rerr)
	}
	defer resp.Body.Close()

	return WebReference{"ok", vars["cname"], true, "http.media.mit.edu", "", 8999}, nil
}

// NB: Move into separate infoblox API library
func listCnames(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	vars := mux.Vars(r)
	log.Printf(" get: %s\n", vars["project"])

	return WebReference{"blessing", "blessing.media.mit.edu", true, "http.media.mit.edu", "", 8999}, nil
}

func listStatus(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	db_handle.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("sites"))
		if buk == nil {
			log.Fatal("bucket not found: sites")
		}
		return nil
	})

	return Status{"ok"}, nil
}

func getProject(project string) (*WebReference, error) {
	aref := WebReference{"gpBlessed", "gpbless.media.mit.edu", true, "http.media.mit.edu", "", 8999}
	db_handle.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("sites"))
		v := buk.Get([]byte(project))
		json.Unmarshal([]byte(v), &aref)
		return nil
	})

	return &aref, nil
}

func ls() ([]string, error) {
	return nil, nil
}

func genPath(project string) string {
	return fmt.Sprintf("%s%s%s.conf", nginxdir, string(os.PathSeparator), project)
}

func writeNginxConf(project string) (err error) {
	aref, _ := getProject(project)
	if settings.VerboseLog {
		log.Printf("will write: %s", aref)
	}

	fullpath := genPath(project)
	f, e := os.OpenFile(fullpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

	if e != nil {
		log.Fatal(e)
	}

	if settings.VerboseLog {
		log.Printf("opened: %s", fullpath)
	}

	io.WriteString(f, "server {\n")
	io.WriteString(f, "  listen 80;\n")
	io.WriteString(f, fmt.Sprintf("  server_name %s;\n", aref.FQDN))
	io.WriteString(f, "  location / {\n")
	io.WriteString(f, "    proxy_set_header Host $host;\n")
	io.WriteString(f, "    proxy_set_header X-Real-IP $remote_addr;\n")
	io.WriteString(f, "    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
	io.WriteString(f, "    proxy_set_header X-Forwarded-Proto $scheme;\n")
	io.WriteString(f, "    add_header X-MIT-Media-Lab \"Oh Yes\";\n")
	io.WriteString(f, fmt.Sprintf("    proxy_pass http://draco.media.mit.edu:%d;\n", aref.Port))
	io.WriteString(f, "    proxy_read_timeout 90;\n")
	io.WriteString(f, "    proxy_http_version 1.1;\n")
	io.WriteString(f, "  }\n")
	io.WriteString(f, "}\n")

	f.Sync()
	f.Close()

	return nil
}

func logByteArray(buf []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		log.Println(scanner.Text())
	}
}

func reloadNginx() error {
	cmd := exec.Command("/usr/sbin/nginx", "-t")

	outb, err := cmd.CombinedOutput()
	logByteArray(outb)
	if err != nil {
		log.Println(err)
		return err
	}

	cmd = exec.Command("/etc/init.d/nginx", "reload")
	outb, err = cmd.CombinedOutput()
	logByteArray(outb)
	if err != nil {
		log.Println(err)
	}

	return err
	/** // useful code snippet for gathering
	    // different filedesc outputs

	cmdstdout, err := cmd.StdoutPipe()
	cmdstderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	outstrbuf := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, cmdstdout)
		outstrbuf <- buf.String()
	}()

	errstrbuf := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, cmdstderr)
		errstrbuf <- buf.String()
	}()

	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
		return err
	}

	outstr := <-outstrbuf
	log.Println(" outstr: '" + outstr + "'\n")
	errstr := <-errstrbuf
	log.Println(" errstr: '" + errstr + "'\n")

	*/
	//var in string
	//	for err != io.EOF {
	//		in, err = stdout.ReadString('\n')
	//		log.Println(in)
	//	}

	//	system("/etc/init.d/nginx reload")
}

func writeSite(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	vars := mux.Vars(r)
	if settings.VerboseLog {
		log.Printf("writeSite: %s\n", vars["project"])
	}
	writeNginxConf(vars["project"])
	err := reloadNginx()
	if err == nil {
	} else {
		log.Println("Will not reload nginx config, moving config file to /tmp/badjuju")
		err = os.Rename(genPath(vars["project"]), "/tmp/badjuju")
		if err != nil {
			log.Println("Trouble moving file")
			log.Println(err)
		}

	}

	return Status{"ok"}, nil
}

func addSite(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	vars := mux.Vars(r)
	if settings.VerboseLog {
		log.Printf("addSite: %s %s\n", vars["project"], vars["cname"])
	}

	db_err := db_handle.Update(func(tx *bolt.Tx) error {
		log.Printf("huh\n")
		aref := WebReference{"default", "blank", false, "http.media.mit.edu", "blank", 8999}
		buk := tx.Bucket([]byte("sites"))
		project := vars["project"]
		aref.Project = project
		aref.FQDN = vars["cname"] + ".media.mit.edu"
		buf, err := json.Marshal(aref)
		if settings.VerboseLog {
			log.Printf("buf: %s\n", buf)
		}
		if err != nil {
			log.Fatal(err)
			return err
		}
		return buk.Put([]byte(aref.Project), buf)
	})

	if db_err != nil {
		log.Fatal(db_err)
	}

	return Status{"ok"}, nil
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func listSites(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	var webrefs []WebReference

	db_handle.View(func(tx *bolt.Tx) error {
		buk := tx.Bucket([]byte("sites"))
		cur := buk.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			aref := WebReference{}
			json.Unmarshal([]byte(v), &aref)
			webrefs = append(webrefs, aref)
		}
		return nil
	})
	return webrefs, nil
}

/*
func openDB() {
	var err error
	db_handle, err = bolt.Open(db_dir_name, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db_handle.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("sites"))
		if err != nil {
			log.Fatal(err)
			return err
		}
		return nil
	})
}
*/

func main() {

	settings.Init()

	app := cli.NewApp()
	app.Usage = "Tool to support Nephelai project creation/deploy"

	app.Version = "ADMETE 0.0.1"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "Verbose, V",
			Usage:       string("Log verbose output to " + app.Name + ".log"),
			Destination: &settings.VerboseLog,
		},
		cli.StringFlag{
			Name:        "config,c",
			Usage:       "Configuration file or directory to load config info",
			Value:       "conf.d/*",
			Destination: &config_info,
		},
		cli.StringFlag{
			Name:        "databasedir,db",
			Usage:       "Database directory for holding okea information",
			Destination: &db_dir_name,
		},
		cli.StringFlag{
			Name:        "docker_endpoint,de",
			Usage:       "End point to connect to docker API.",
			Value:       "tcp://127.0.0.1:2375",
			Destination: &docker_endpoint,
		},
		cli.StringFlag{
			Name:        "nginxdir,n",
			Usage:       "Location of nginx configuration directory for live sites",
			Destination: &nginxdir,
		},
		cli.StringFlag{
			Name:        "nginxconf",
			Usage:       "YAML file configuration used by Ansible to reconfigure nginx",
			Destination: &nginx_yaml_file_name,
		},
		cli.IntFlag{
			Name:        "port,p",
			Usage:       "Port to listen on",
			Value:       12098,
			Destination: &listen_port,
		},
		cli.StringFlag{
			Name:        "ip,i",
			Usage:       "IP Address to bind to",
			Value:       "0.0.0.0",
			Destination: &listen_ipaddress,
		},
		cli.StringFlag{
			Name:        "api_auth_infoblox,a",
			Usage:       "Connection token string for infoblox, format: <user>:<password>",
			Value:       "<user>:<password>",
			Destination: &auth_token,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "status",
			Usage:       "okea status",
			Description: "show status information about okea",
			Action: func(c *cli.Context) {
				println("check ", c.Args().First())
				services.ReloadNGINXConfig(docker_endpoint)
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
	}

	// This works effectively as the main of the program
	app.Action = func(c *cli.Context) {
		var logOut io.Writer
		var err error
		if settings.VerboseLog {
			log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
			logOut, err = os.OpenFile(string(app.Name+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf(string("Unable to open file: " + app.Name + ".log"))
			}
			log.SetOutput(logOut)
			log.Println("------- Starting " + app.Name + " -------")

		}

		// nginx_conf, _ := config.ReadNGINXConfig(nginx_yaml_file_name)
		//log.Printf("  parsed nginx yaml: %v\n", nginx_conf)
		mdb, _ := db.OpenDBs(db_dir_name)
		log.Printf("  DBs opened: %v\n", mdb)

		/*
			router := mux.NewRouter()
			router.Handle("/cnames", handler(listCnames)).Methods("GET")
			router.Handle("/cnames/{project}", handler(listCnames)).Methods("GET")
			router.Handle("/cnames/{project}/{cname}", handler(updateCname)).Methods("POST")
			router.Handle("/cnames/{id}", handler(listCnames)).Methods("DELETE")
			router.Handle("/cnames/{project}", handler(listCnames)).Methods("DELETE")

			router.Handle("/status", handler(listStatus)).Methods("GET")

			router.Handle("/writesite/{project}", handler(writeSite)).Methods("POST")
			router.Handle("/sites/{project}/{cname}", handler(addSite)).Methods("POST")
			router.Handle("/sites", handler(listSites)).Methods("GET")

			http.Handle("/", router)

		*/

		router := routers.InitRoutes()
		n := negroni.New(negroni.NewRecovery(), core.NewLogger(logOut))
		n.UseHandler(router)

		log.Printf("Listening on %s:%d\n", listen_ipaddress, listen_port)
		addr := fmt.Sprintf("%s:%d", listen_ipaddress, listen_port)
		err = http.ListenAndServeTLS(addr, "conf.d/ml-cert.pem", "conf.d/ml-key.pem", n)

		log.Println(err.Error())
	}

	app.Run(os.Args)
}
