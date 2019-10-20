package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/howeyc/fsnotify"
)

var operationKeywords map[string]string

func init() {
	operationKeywords = map[string]string{
		"move":   "mv",
		"copy":   "cp",
		"delete": "rm",
	}
}

func main() {

	var operationType string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "-t	-type specifies the kind of operation to run, values are: move, copy, delete")
	}

	flag.StringVar(&operationType, "type", "none", "specifies what kind of operation to run possible values are move, copy, delete")
	flag.StringVar(&operationType, "t", "none", "specifies what kind of operation to run possible values are move, copy, delete")

	flag.Parse()

	if _, ok := operationKeywords[operationType]; !ok {
		log.Println("Possible values for the -t -type flag are: move, copy, delete")
		os.Exit(1)
	}

	configFile, err := getConfigFile()
	if err != nil {
		log.Printf("Error with config.json file: %v", err)
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Println("couldn't unmarshal config.json file", err)
	}

	watch(operationType, config)
}

// getConfigFile will look for a config.json file from where the command is being called
func getConfigFile() ([]byte, error) {
	cmd := exec.Command("pwd")
	dir, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	configPath := strings.TrimSpace(string(dir)) + "/config.json"

	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.New("Couldn't read config.json file")
	}

	return configFile, nil
}

func watch(operation string, config Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}

	defer watcher.Close()
	done := make(chan bool)
	runner := false
	go func() {
		if !runner {
			do(config, operation)
		}
		for {
			select {
			case event, ok := <-watcher.Event:
				if !ok {
					return
				}
				if event.IsCreate() {
					do(config, operation)
				}
			case err, ok := <-watcher.Error:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	paths := getConfigPaths(config)

	// start watching all the paths concurrently
	for _, i := range paths {
		go func(path string) {
			err = watcher.Watch(path)
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}
	<-done
}

// do finds the files in the watch folder and starts performing the operation
func do(config Config, op string) error {

	for _, i := range config.File {
		var foundFiles []string

		folder, err := os.Open(i.Watch)
		if err != nil {
			return errors.New("couldn't open watch folder \n:" + err.Error())
		}

		files, err := folder.Readdirnames(0)
		if err != nil {
			return errors.New("couldn't read from watch folder \n:" + err.Error())
		}

		for _, extension := range i.Extensions {
			for _, fileName := range files {
				if strings.HasSuffix(fileName, extension) {
					foundFiles = append(foundFiles, fileName)
				}
			}
		}

		err = operation(op, i.Watch, i.Destination, foundFiles)
		if err != nil {
			return errors.New("couldn't run intended operation \n" + err.Error())
		}
	}
	return nil
}

func getConfigPaths(config Config) []string {
	var watchPaths []string

	for _, i := range config.File {
		watchPaths = append(watchPaths, i.Watch)
	}
	return watchPaths
}

func operation(op, oldPath, newPath string, files []string) error {
	operationType := operationKeywords[op]

	var err error
	if operationType == "mv" || operationType == "cp" {

		for _, singleFile := range files {
			oldFilePath := oldPath + "/" + singleFile

			cmd := exec.Command(operationType, oldFilePath, newPath)
			err = cmd.Run()
		}
	} else if operationType == "rm" {

		for _, singleFile := range files {
			oldFilePath := oldPath + "/" + singleFile

			cmd := exec.Command(operationType, oldFilePath)
			err = cmd.Run()
		}

	}
	return err
}
