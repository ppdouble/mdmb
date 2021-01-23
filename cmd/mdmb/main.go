package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"

	"github.com/jessepeterson/mdmb/internal/device"
	"github.com/jessepeterson/mdmb/internal/mdmclient"
)

type subCmdFn func(string, []string, func())

type subCmd struct {
	Name        string
	Description string
	Func        subCmdFn
}

var subCmds []subCmd = []subCmd{
	{"help", "Display usage help", help},
	{"enroll", "enroll devices into MDM", enroll},
}

func help(_ string, _ []string, usage func()) {
	usage()
}

func main() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// var (
	// 	dbPath = f.String("db", "mdmb.db", "mdmb database file path")
	// )
	f.Usage = func() {
		fmt.Fprintf(f.Output(), "%s [flags] <subcommand> [flags]\n", f.Name())
		fmt.Fprint(f.Output(), "\nFlags:\n")
		f.PrintDefaults()
		fmt.Fprint(f.Output(), "\nSubcommands:\n")
		for _, sc := range subCmds {
			fmt.Fprintf(f.Output(), "    %s\t%s\n", sc.Name, sc.Description)
		}
	}
	f.Parse(os.Args[1:])

	if len(f.Args()) < 1 {
		fmt.Fprintln(f.Output(), "no subcommand supplied")
		f.Usage()
		os.Exit(2)
	}

	for _, sc := range subCmds {
		if f.Args()[0] == sc.Name {
			sc.Func(sc.Name, f.Args()[1:], f.Usage)
			return
		}
	}

	fmt.Fprintf(f.Output(), "invalid subcommand: %s\n", f.Args()[0])
	f.Usage()
	os.Exit(2)
}

func enroll(name string, args []string, usage func()) {
	f := flag.NewFlagSet(name, flag.ExitOnError)
	var (
		// enrollType = f.String("type", "profile", "enrollment type")
		// number     = f.Int("n", 1, "number of devices")
		url  = f.String("url", "", "URL pointing to enrollment spec (e.g. profile)")
		file = f.String("file", "", "file of enrollment spec (e.g. profile)")
	)
	f.Usage = func() {
		usage()
		fmt.Fprintf(f.Output(), "\nFlags for %s subcommand:\n", f.Name())
		f.PrintDefaults()
	}
	f.Parse(args)

	if (*url == "" && *file == "") || (*url != "" && *file != "") {
		fmt.Fprintln(f.Output(), "must specify one enrollment url or file")
		f.Usage()
		os.Exit(2)
	}

	if *url != "" {
		fmt.Fprintln(f.Output(), "-url not yet supported")
		os.Exit(1)
	}

	if err := enrollWithFile(*file); err != nil {
		stdlog.Fatal(err)
	}

	// c := client.NewMDMClient()
	// fmt.Println(c.UDID)
}

func enrollWithFile(path string) error {

	ep, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	dev := &device.Device{
		UDID:         "475F0A29-6FCE-419E-A30F-9FF616FD2B87",
		Serial:       "P3IJDS49Z90A",
		ComputerName: "Malik's computer",
	}

	client := mdmclient.NewMDMClient(dev)

	err = client.Enroll(ep, rand.Reader)
	if err != nil {
		return err
	}

	err = client.Connect()
	if err != nil {
		return err
	}

	return nil
}
