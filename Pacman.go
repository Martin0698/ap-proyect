package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
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
var player sprite
var ghosts []*sprite
var score int
var numDots int
var lives = 1

//LOAD MAZE

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

	//CAPTURE PLAYER POSITION
	// TRAVERSE EACH CHARACTER OF THE MAZE AND CREATE
	// A NEW PLAYER WHEN IT LOCATES A 'P'
	for row, line := range maze {
		for col, char := range line {
			switch char {
			case 'P':
				player = sprite{row, col}
			case 'G':
				ghosts = append(ghosts, &sprite{row, col})
			case '.':
				numDots++
			}
		}
	}

	return nil
}

//2.  PRINTING TO THE TERMINAL

func printMaze() {
	ClearScreen()
	//It does not working
	//we need to clear the screen after each loop
	for _, line := range maze {
		for _, chr := range line {
			switch chr {
			case '#':
				fallthrough
			case '.':
				fmt.Printf("%c", chr)
			default:
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	MoveCursor(player.row, player.col)
	fmt.Print("P")

	for _, g := range ghosts {
		MoveCursor(g.row, g.col)
		fmt.Print("G")
	}

	MoveCursor(len(maze)+1, 0)

	//PRINTING SCORE
	fmt.Println("Score", score, "\t Lives: ", lives)
}

//3. GAME LOOP
// MAIN
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
		//process movement
		movePlayer(input)
		moveGhosts()

		//process collisions
		for _, g := range ghosts {
			if player == *g {
				lives = 0
			}
		}

		// Game Over Cases:
		if input == "ESC" || numDots == 0 || lives <= 0 {
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
	cbTerm := exec.Command("stty", "cbreak", "-echo")
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

		// KEY PRESSES
	} else if cnt >= 3 {
		if buffer[0] == 0x1b && buffer[1] == '[' {
			switch buffer[2] {
			case 'A':
				return "UP", nil
			case 'B':
				return "DOWN", nil
			case 'C':
				return "RIGHT", nil
			case 'D':
				return "LEFT", nil
			}

		}
	}
	return "", nil
}

// 5. MOVEMENT

// TRacking player position

type sprite struct {
	row int
	col int
}

//HANDLE MOVEMENT

func makeMove(oldRow, oldCol int, dir string) (newRow, newCol int) {
	newRow, newCol = oldRow, oldCol

	switch dir {
	case "UP":
		newRow = newRow - 1
		if newRow < 0 {
			newRow = len(maze) - 1
		}
	case "DOWN":
		newRow = newRow + 1
		if newRow == len(maze) {
			newRow = 0
		}
	case "RIGHT":
		newCol = newCol + 1
		if newCol == len(maze[0]) {
			newCol = 0
		}
	case "LEFT":
		newCol = newCol - 1
		if newCol < 0 {
			newCol = len(maze[0]) - 1
		}
	}
	//if new position hits a wall '#'
	//movement is cancelled
	if maze[newRow][newCol] == '#' {
		newRow = oldRow
		newCol = oldCol
	}
	return
}

//Move player
func movePlayer(dir string) {
	player.row, player.col = makeMove(player.row, player.col, dir)
	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		//Remove dot from the maze
		maze[player.row] = maze[player.row][0:player.col] + " " + maze[player.row][player.col+1:]

	}
}

//4.  GHOSTS

////////////////////////////////////
////							////
////		MAKING GHOSTS		////
///			    				////
////////////////////////////////////

//Random generator to control our ghosts.
func drawDirection() string {
	dir := rand.Intn(4)
	move := map[int]string{
		0: "UP",
		1: "DOWN",
		2: "RIGHT",
		3: "LEFT",
	}
	return move[dir]
}

func moveGhosts() {
	for _, g := range ghosts {
		dir := drawDirection()
		g.row, g.col = makeMove(g.row, g.col, dir)
	}
}

// 5. Game win

// ClearScreen cleans the terminal and set cursor position to the top left corner.
func ClearScreen() {
	fmt.Print("\x1b[2J")
	MoveCursor(0, 0)
}

// MoveCursor sets the cursor position to given row and col.
//
// Please note that ANSI is 1-based and the top left corner is (1,1), but here we are assuming
// the user is using a zero based coordinate system where the top left corner is (0, 0)
func MoveCursor(row, col int) {
	fmt.Printf("\x1b[%d;%df", row+1, col+1)
}

const reset = "\x1b[0m"

type Colour int

const (
	BLACK Colour = iota
	RED
	GREEN
	BROWN
	BLUE
	MAGENTA
	CYAN
	GREY
)

var colours = map[Colour]string{
	BLACK:   "\x1b[1;30;40m",
	RED:     "\x1b[1;31;41m",
	GREEN:   "\x1b[1;32;42m",
	BROWN:   "\x1b[1;33;43m",
	BLUE:    "\x1b[1;34;44m",
	MAGENTA: "\x1b[1;35;45m",
	CYAN:    "\x1b[1;36;46m",
	GREY:    "\x1b[1;37;47m",
}

// WithBlueBackground wraps the given text with blue background and reset escape sequences.
func WithBlueBackground(text string) string {
	return "\x1b[44m" + text + reset
}

// WithBackground wraps the given 'text' with 'colour' background and reset escape sequences.
func WithBackground(text string, colour Colour) string {
	if c, ok := colours[colour]; ok {
		return c + text + reset
	}
	//Default to blue if none resolved
	return WithBlueBackground(text)
}
