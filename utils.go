package nuke

import (
	"log"
	"runtime/debug"
)

func Background(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in background function: %v\n%s", r, debug.Stack())
			}
		}()
		f()
	}()
}
