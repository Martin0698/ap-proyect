package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

//READ THE MAZE DATA
/*
- # represents a wall
- . represents a dot
- P represents the player
- G represents the ghosts
- X represents the power up pills
*/

// 7. EMOJIS, CONFIG JSONS

//Struct tag
type config struct {
	Player    string        `json:"player"`
	Ghost     string        `json:"ghost"`
	Wall      string        `json:"wall"`
	Dot       string        `json:"dot"`
	Pill      string        `json:"pill"`
	Death     string        `json:"death"`
	Space     string        `json:"space"`
	UseEmoji  bool          `json:"use_emoji"`
	GhostBlue string        `json:"ghost_blue"`
	PillTime  time.Duration `json:"pillTime"`
}

type GhostStatus string

type ghost struct {
	position sprite
	status   GhostStatus
}

const (
	GhostStatusNormal GhostStatus = "Normal"
	GhostStatusBlue   GhostStatus = "Blue"
)

// 1. Read the file maze01.txt

var maze []string
var maze02 []string
var player sprite
var ghosts []*ghost
var score int
var numDots int
var lives = 3
var cfg config
var pillTimer *time.Timer
var ghostsStatusMx sync.RWMutex
var pillMx sync.Mutex

// Load Configuration from jsons files

func loadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

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
				player = sprite{row, col, row, col}
			case 'G':
				ghosts = append(ghosts, &ghost{sprite{row, col, row, col}, GhostStatusNormal})
			case '.':
				numDots++
			}
		}
	}

	return nil
}

func MoveEmoji(row, col int) {
	if cfg.UseEmoji {
		MoveCursor(row, col*2)
	} else {
		MoveCursor(row, col)
	}
}

func processPill() {
	pillMx.Lock()
	updateGhosts(ghosts, GhostStatusBlue)
	if pillTimer != nil {
		pillTimer.Stop()
	}
	pillTimer = time.NewTimer(time.Second * cfg.PillTime)
	pillMx.Unlock()
	<-pillTimer.C
	pillMx.Lock()
	pillTimer.Stop()
	updateGhosts(ghosts, GhostStatusNormal)
	pillMx.Unlock()
}

//2.  PRINTING TO THE TERMINAL
func printMaze() {
	ClearScreen()

	//we need to clear the screen after each loop
	for _, line := range maze {
		for _, chr := range line {
			switch chr {
			case '#':
				fmt.Print(WithBlueBackground(cfg.Wall))
			case '.':
				fmt.Printf(cfg.Dot)
			case 'X':
				fmt.Print(cfg.Pill)
			default:
				fmt.Print(cfg.Space)
			}
		}
		fmt.Println()
	}
	MoveEmoji(player.row, player.col)
	fmt.Print(cfg.Player)

	ghostsStatusMx.RLock()

	for _, g := range ghosts {
		MoveEmoji(g.position.row, g.position.col)
		if g.status == GhostStatusNormal {
			fmt.Printf(cfg.Ghost)
		} else if g.status == GhostStatusBlue {
			fmt.Printf(cfg.GhostBlue)
		}

	}
	ghostsStatusMx.RUnlock()

	MoveEmoji(len(maze)+1, 0)

	livesRemaining := strconv.Itoa(lives)
	if cfg.UseEmoji {
		livesRemaining = getLivesAsEmoji()
	}

	//PRINTING SCORE
	fmt.Println("Score", score, "\t Lives: ", livesRemaining)
}

func getLivesAsEmoji() string {
	buf := bytes.Buffer{}
	for i := lives; i > 0; i-- {
		buf.WriteString(cfg.Player)
	}
	return buf.String()
}

func updateGhosts(ghost []*ghost, ghostStatus GhostStatus) {
	ghostsStatusMx.Lock()
	defer ghostsStatusMx.Unlock()
	for _, g := range ghosts {
		g.status = ghostStatus
	}
}

//3. GAME LOOP
// MAIN

func main() {
	flag.Parse()

	fmt.Println("Enter the numbers of ghosts: ")
	var num string
	fmt.Scanln(&num)

	cbreakMode()
	defer cookedMode()

	if num == "1" {
		//Load Maze
		err := loadMaze(*mazeFile)
		if err != nil {
			log.Println("failed to load maze: ", err)
			return

		}
		err = loadConfig(*configFile)
		if err != nil {
			log.Println("Failed to load configuration", err)
			return
		}

	} else if num == "2" {
		err := loadMaze(*mazeFile2)
		if err != nil {
			log.Println("Failed to load configuration", err)
			return
		}
		err = loadConfig(*configFile)
		if err != nil {
			log.Println("Failed to load configuration", err)
			return
		}
	}

	//process input async way
	input := make(chan string)
	go func(ch chan<- string) {
		for {
			input, err := readInput()
			if err != nil {
				log.Println("error reading input:", err)
				ch <- "ESC"
			}
			ch <- input
		}
	}(input)
	for {

		select {
		case inp := <-input:
			if inp == "ESC" {
				lives = 0
			}
			movePlayer(inp)
		default:
		}
		//process movement

		moveGhosts()

		//process collisions
		for _, g := range ghosts {
			if player.row == g.position.row && player.col == g.position.col {
				ghostsStatusMx.RLock()
				if g.status == GhostStatusNormal {
					lives = lives - 1
					if lives != 0 {
						MoveEmoji(player.row, player.col)
						fmt.Print(cfg.Death)
						MoveEmoji(len(maze)+2, 0)
						ghostsStatusMx.RUnlock()
						updateGhosts(ghosts, GhostStatusNormal)
						time.Sleep(1000 * time.Millisecond)
						player.row, player.col = player.startRow, player.startCol
					}
				} else if g.status == GhostStatusBlue {
					ghostsStatusMx.RUnlock()
					updateGhosts([]*ghost{g}, GhostStatusNormal)
					g.position.row, g.position.col = g.position.startRow, g.position.startCol
				}
			}
		}

		// update screen
		printMaze()

		// Game Over Cases:
		if numDots == 0 || lives <= 0 {
			if lives == 0 {
				MoveEmoji(player.row, player.col)
				fmt.Print(cfg.Death)
				MoveCursor(player.startRow, player.startCol-1)
				fmt.Print("GAME OVER")
				MoveEmoji(len(maze)+2, 0)
			}
			break
		}
		// repeat
		time.Sleep(200 * time.Millisecond)
	}
}

//4. TERMINAL MODE -> CBREAK MODE

/*
 cbreak mode some characters are preprocessed and
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

// Tracking player position

type sprite struct {
	row      int
	col      int
	startRow int
	startCol int
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
		if newRow == len(maze)-1 {
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

	removeDot := func(row, col int) {
		maze[row] = maze[row][0:col] + " " + maze[row][col+1:]
	}

	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		//Remove dot from the maze
		//maze[player.row] = maze[player.row][0:player.col] + " " + maze[player.row][player.col+1:]

		removeDot(player.row, player.col)
	case 'X':
		score += 10
		removeDot(player.row, player.col)
		go processPill()
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
		g.position.row, g.position.col = makeMove(g.position.row, g.position.col, dir)
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

var (
	configFile = flag.String("config-file", "config.json", "path to custom configuration file")
	mazeFile   = flag.String("maze-file", "maze01.txt", "path to a custom maze file")
	mazeFile2  = flag.String("maze-file02", "maze02.txt", "path to a custom a maze file")
)
