package main

import (
	"fmt"
	"os"

	"golang.org/x/tools/cover"
)

func main() {
	filename := os.Args[1]
	profiles, err := cover.ParseProfiles(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(profiles))

	byFile := make(map[string]*[]cover.ProfileBlock, len(profiles))
	for _, profile := range profiles {
		fmt.Println(profile.FileName)
		if blocks, found := byFile[profile.FileName]; found {
			*blocks = append(*blocks, profile.Blocks...)
		} else {
			byFile[profile.FileName] = &profile.Blocks
		}
	}
	fmt.Println(len(byFile))
}
