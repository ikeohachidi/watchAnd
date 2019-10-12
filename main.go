package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/howeyc/fsnotify"
)

// todo: add support for copy and delete
// todo: change into a command line code

func main() {
	configFile, err := getConfigFile()
	if err != nil {
		log.Println("couldn't work with config file", err)
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Println("couldn't unmarshal configFile", err)
	}

	err = watchAndDo(config)
	if err != nil {
		log.Println("couldn't watch and do", err)
	}
}

// getConfigFile will look for a config file from where the command is being called
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

func watch(path, destination string, files []string, fn operationFunc) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}

	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Event:
				if !ok {
					return
				}
				if event.IsCreate() {
					fn(path, destination, files)
				}
			case err, ok := <-watcher.Error:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Watch(path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func watchAndDo(config Config) error {
	// watch the folder in config.file.watch
	// if file exists in there send it to destination

	done := make(chan string)
	var codeErr error

	for _, i := range config.File {

		go func(configOp FileStruct) {
			var foundFiles []string

			folder, err := os.Open(configOp.Watch)
			if err != nil {
				log.Printf("couldn't open watch folder \n%v", err)
				os.Exit(1)
			}

			files, err := folder.Readdirnames(0)
			if err != nil {
				log.Printf("couldn't read from watch folder \n%v", err)
				os.Exit(1)
			}

			for _, extension := range configOp.Extensions {
				for _, fileName := range files {
					if strings.HasSuffix(fileName, extension) {
						foundFiles = append(foundFiles, fileName)
					}
				}
			}

			// run code once and then start running watch
			err = operation(configOp.Watch, configOp.Destination, foundFiles)
			if err != nil {
				log.Printf("couldn't run intended operation \n%v", err)
				os.Exit(1)
			}

			watch(configOp.Watch, configOp.Destination, foundFiles, operation)

		}(i)

	}
	<-done
	return codeErr
}

func operation(oldPath, newPath string, files []string) error {
	for _, singleFile := range files {
		oldFilePath := oldPath + "/" + singleFile

		cmd := exec.Command("mv", oldFilePath, newPath)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
