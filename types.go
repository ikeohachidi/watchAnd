package main

// Config is the struct which maps to the expected fields for the config file used
type Config struct {
	File []FileStruct `json:"file"`
}

// FileStruct provides the Config.File structure
type FileStruct struct {
	Extensions  []string `json:"extensions"`
	Watch       string   `json:"watch"`
	Destination string   `json:"destination"`
}

type operationFunc func(path, destination string, files []string) error
