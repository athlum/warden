package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.elenet.me/qi.feng/warden"
	"git.elenet.me/qi.feng/warden/backends"
	"log"
	"os"
	"text/template"
)

type CmdLineOpts struct {
	data         string
	templatePath string
	dest         string
}

var (
	opts   CmdLineOpts
	config *warden.GuardianConfig
)

func init() {
	flag.StringVar(&opts.data, "data", "", "App id noted on zk.")
	flag.StringVar(&opts.templatePath, "templ", "", "Template file for render.")
	flag.StringVar(&opts.dest, "dest", "", "Render target.")
}

func main() {
	flag.Parse()
	log.Printf("args: %+v", opts)
	if opts.data == "" || opts.templatePath == "" || opts.dest == "" {
		log.Println("Appid, templ and dest could not be null.")
		os.Exit(0)
	}

	templ, tempErr := template.ParseFiles(opts.templatePath)
	if tempErr != nil {
		log.Println("Loading template file failed.")
		os.Exit(1)
	}

	if _, err := os.Stat(opts.dest); err == nil {
		err = os.Rename(opts.dest, fmt.Sprintf("%s.bak", opts.dest))
		if err != nil {
			log.Println(err)
			log.Printf("%s existed, failed on backup", opts.dest)
			os.Exit(1)
		}
	}
	f, fileErr := os.Create(opts.dest)
	if fileErr != nil {
		log.Fatal(fileErr)
	}

	appNodes := []backends.AppNode{}
	if err := json.Unmarshal([]byte(opts.data), &appNodes); err != nil {
		log.Fatal(err)
	}

	if err := templ.Execute(f, appNodes); err != nil {
		log.Fatal(err)
	}
}
