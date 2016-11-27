package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/prologic/httpfs/fsapi"
)

// debug flag enables logging of debug messages to stderr.
var debug = flag.Bool("debug", false, "enable debug log messages to stderr")
var url = flag.String("url", "", "url of httpsfs backend (required)")
var tlsverify = flag.Bool("tlsverify", false, "enable TLS verification")
var mount = flag.String("mount", "", "path to mount volume (required)")

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func debugLog(msg interface{}) {
	fmt.Printf("%s\n", msg)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	if *mount == "" || *url == "" {
		usage()
		os.Exit(2)
	}

	c, err := fuse.Mount(
		*mount,
		fuse.FSName("httpfs"),
		fuse.Subtype("httpfs"),
		fuse.VolumeName("HTTP FS"),
		// fuse.LocalVolume(),
		fuse.AllowOther(),

		fuse.MaxReadahead(2^20),
		fuse.NoAppleDouble(),
		fuse.NoAppleXattr(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	cfg := &fs.Config{}
	if *debug {
		cfg.Debug = debugLog
	}
	srv := fs.New(c, cfg)
	filesys := fsapi.NewHTTPFS(*url, *tlsverify)

	if err := srv.Serve(filesys); err != nil {
		log.Fatal(err)
	}

	// Check if the mount process has an error to report.
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}
