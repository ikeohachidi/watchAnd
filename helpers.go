package main

import "log"

func checkErrorPrint(err error, message ...string) {
	if err != nil {
		log.Println(err, message)
	}
}
