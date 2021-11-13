package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

//READ THE MAZE DATA
/*
- # represents a wall
- . represents a dot
- P represents the player
- G represents the ghosts
- X represents the power up pills
*/

// 1. Read the file maze01.txt

var maze []string

func loadMaze(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		maze = append(maze, line)
	}

	return nil
}

//2.  PRINTING TO THE TERMINAL

func printMaze() {
	for _, line := range maze {
		fmt.Println(line)
	}
}

//3. GAME LOOP

func main() {
	err := loadMaze("maze01.txt")
	if err != nil {
		log.Println("failed to load maze: ", err)
		return
	}
	for {
		printMaze()
		break
	}
}
