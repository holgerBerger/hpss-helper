package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"regexp"
)

func listarchives() {
	matches, err := filepath.Glob(config.General.Cachedir + "/*.full*")
	if err != nil {
		log.Fatal("could not list archives", err)
	}
	if len(matches) > 0 {
		fmt.Println("available archives:")
	} else {
		fmt.Println("no archives available")
	}
	for _, match := range matches {
		re := regexp.MustCompile(`^(.*)\.(full.*)$`)
		m := re.FindStringSubmatch(path.Base(match))
		if m[2] == "full" {
			fmt.Printf("  %-40s (index cached) \n", m[1])
		} else {
			fmt.Printf("  %-40s (index offline) \n", m[1])
		}
	}
}
