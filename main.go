package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/storj"
)

type Config struct {
	Key     uplink.APIKey
	Link    *uplink.Uplink
	Project *uplink.Project
	Bucket  *uplink.Bucket

	SatelliteAddr string
	EncryptionKey storj.Key
	Path          storj.Path
	File          string
}

type Status struct {
	Time    time.Time
	Success bool
	Error   string
}

type Stats struct {
	List       map[int64]Status
	File       string
	Size       int64
	Interval   time.Duration
	Percentage float32
}

var (
	Lock *sync.Mutex
	Stat Stats
	tmpl *template.Template
)

func toStorjKey(key string) (newKey storj.Key) {
	var encryptionKey storj.Key
	copy(encryptionKey[:], []byte(key))
	return encryptionKey
}

func parseApiKey(key string) (apiKey uplink.APIKey, err error) {
	apiKey, err = uplink.ParseAPIKey(key)
	if err != nil {
		return uplink.APIKey{}, err
	}
	return
}

func setupUplink() (link *uplink.Uplink, err error) {
	ctx := context.Background()

	// Create an Uplink object with a default config
	link, err = uplink.NewUplink(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create new Uplink object: %v", err)
	}
	return
}

func openProject(ctx context.Context, conf Config) (proj *uplink.Project, err error) {
	opts := uplink.ProjectOptions{}
	opts.Volatile.EncryptionKey = &conf.EncryptionKey

	proj, err = conf.Link.OpenProject(ctx, conf.SatelliteAddr, conf.Key, &opts)
	if err != nil {
		return &uplink.Project{}, fmt.Errorf("could not open project: %v", err)
	}
	return proj, nil
}

func openBucket(ctx context.Context, conf Config, name string) (bucket *uplink.Bucket, err error) {
	bucket, err = conf.Project.OpenBucket(ctx, name, &uplink.EncryptionAccess{Key: conf.EncryptionKey})
	if err != nil {
		return &uplink.Bucket{}, fmt.Errorf("could not open bucket %q: %v", name, err)
	}
	return bucket, nil
}

func uploadData(ctx context.Context, conf *Config, data io.Reader) (err error) {
	log.Printf("uploading file: %v", conf.Path)

	err = conf.Bucket.UploadObject(ctx, conf.Path, data, nil)
	if err != nil {
		return fmt.Errorf("could not upload: %v", err)
	}
	return nil
}

func downloadDataandCompare(ctx context.Context, conf *Config) (err error) {

	// Initiate a download of the same object again
	readBack, err := conf.Bucket.OpenObject(ctx, conf.Path)
	if err != nil {
		return fmt.Errorf("could not open object at %q: %v", conf.Path, err)
	}
	defer readBack.Close()

	// We want the whole thing, so range from 0 to -1
	strm, err := readBack.DownloadRange(ctx, 0, -1)
	if err != nil {
		return fmt.Errorf("could not initiate download: %v", err)
	}
	defer strm.Close()

	// Read everything from the stream
	receivedContents, err := ioutil.ReadAll(strm)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	orgcheck, err := ioutil.ReadFile(conf.File)

	if !bytes.Equal(receivedContents, orgcheck) {
		return fmt.Errorf("got different object back: %q != %q", orgcheck, receivedContents)
	}
	return nil
}

func durabilityCheck(ctx context.Context, conf *Config, intv time.Duration) {
	go func() {
		var i, j int64
		var err error
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(intv)
		active := true

		Stat.Interval = intv

		for active {
			select {
			case <-ticker.C:
				i++
				log.Println("checking the file..")
				stat := Status{Time: time.Now().UTC()}
				err = downloadDataandCompare(ctx, conf)
				if err != nil {
					stat.Success = false
					stat.Error = err.Error()
					Lock.Lock()
					Stat.Percentage = float32(j) * 100 / float32(i)
					Stat.List[i] = stat
					Lock.Unlock()
					log.Printf("durability check failed: %v", err)
				} else {
					j++
					stat.Success = true
					stat.Error = "-"
					Lock.Lock()
					Stat.Percentage = float32(j) * 100 / float32(i)
					Stat.List[i] = stat
					Lock.Unlock()
					log.Println("durability check successful")
				}
			case <-quit:
				log.Printf("got SIGTERM, shutting down")
				ticker.Stop()
				active = false

				// Shutdown System
				err = shutdown(conf)
				if err != nil {
					log.Fatal(err)
				}
				os.Exit(0)
			}
		}
	}()
}

func shutdown(conf *Config) (err error) {
	err = conf.Bucket.Close()
	if err != nil {
		return err
	}
	err = conf.Project.Close()
	if err != nil {
		return err
	}
	err = conf.Link.Close()
	if err != nil {
		return err
	}
	return nil
}

func durabilityStats(w http.ResponseWriter, r *http.Request) {

	Lock.Lock()
	err := tmpl.Execute(w, Stat)
	Lock.Unlock()

	if err != nil {
		log.Printf("error encoding stats: %v", err)
	}
}

func main() {
	var running Config
	var err error

	addr := flag.String("addr", "satellite.stefan-benten.de:7777", "Satellite Address")
	apikey := flag.String("apikey", "", "API Key")
	enckey := flag.String("enckey", "you'll never guess this", "Encryption Key")
	bucket := flag.String("bucket", "file-durability", "Bucket")
	path := flag.String("path", "", "path for the file in the bucket")
	file := flag.String("file", "", "File to use for Durability Test")
	interval := flag.Int("interval", 300, "time in seconds between durability check")
	listen := flag.String("listen", ":8080", "Webserver Listen Address")
	flag.Parse()

	if *addr == "" || *apikey == "" || *file == "" {
		log.Fatalln("required flag missing!")
	}

	ctx := context.Background()

	running.Key, err = parseApiKey(*apikey)
	if err != nil {
		log.Fatalln("failed parse api key: ", err)
	}

	running.EncryptionKey = toStorjKey(*enckey)
	running.SatelliteAddr = *addr
	running.File = *file

	running.Link, err = setupUplink()
	if err != nil {
		log.Fatalln("failed to setup uplink: ", err)
	}

	running.Project, err = openProject(ctx, running)
	if err != nil {
		log.Fatalln("failed to open project: ", err)
	}

	//TODO: Check if bucket exists, else create
	running.Bucket, err = openBucket(ctx, running, *bucket)
	if err != nil {
		log.Fatalln("failed to open bucket: ", err)
	}

	openpath, err := filepath.Abs(running.File)
	if err != nil {
		log.Fatalln("could not locate file: ", err)
	}

	testfile, err := os.Open(openpath)
	if err != nil {
		log.Fatalln("failed to open file: ", err)
	}

	fileInfo, err := testfile.Stat()
	if err != nil {
		log.Fatalln("failed to gather file information: ", err)
	}

	if *path != "" {
		running.Path = storj.JoinPaths(*path, running.File)
	} else {
		running.Path = running.File
	}

	err = uploadData(ctx, &running, testfile)
	if err != nil {
		log.Fatalln("failed to upload data: ", err)
	}

	Lock = &sync.Mutex{}
	Stat.List = make(map[int64]Status)
	Stat.Percentage = 100
	Stat.File = fileInfo.Name()
	Stat.Size = fileInfo.Size()

	durabilityCheck(ctx, &running, time.Second*time.Duration(*interval))

	// Creating template from inline file, to make binary portable
	tmpl = template.New("index")
	tmpl, err = tmpl.Parse(indexTemplate)
	if err != nil {
		log.Fatalln("failed to parse template: ", err)
	}

	log.Println("starting web server")
	http.HandleFunc("/", durabilityStats)
	if err := http.ListenAndServe(*listen, nil); err != nil {
		panic(err)
	}
}
