package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	//simpleansi.ClearScreen() //we need to clear the screen after each loop
	for _, line := range maze {
		fmt.Println(line)
	}
}

//3. GAME LOOP

func main() {

	cbreakMode()
	defer cookedMode()

	err := loadMaze("maze01.txt")
	if err != nil {
		log.Println("failed to load maze: ", err)
		return
	}
	for {
		printMaze()

		input, err := readInput()
		if err != nil {
			log.Println("Error reading input: ", err)
			break
		}

		if input == "ESC" {
			break
		}
	}
}

//4. TERMINAL MODE -> CBREAK MODE

/*
 cbreak mode somo characters are preprocessed and
 some are not. This terminal mode to allow us
 to handle the keys "esc" and arrow keys
*/

func cbreakMode() {
	cbTerm := exec.Command("stty", "cbreak", "echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		log.Fatalln("Unable to activate cbreak mode: ", err)
	}
}

//RESTORING COOKED MODE

func cookedMode() {
	cookedTerm := exec.Command("stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalln("Unable to restore cooked mode: ", err)
	}
}

//Reading from StdIn

func readInput() (string, error) {
	buffer := make([]byte, 100)

	cnt, err := os.Stdin.Read(buffer)
	if err != nil {
		return "", err
	}
	if cnt == 1 && buffer[0] == 0x1b { //Ox1b represents ESCAPE key
		return "ESC", nil
	}
	return "", nil
}
