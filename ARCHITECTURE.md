# Pacman Architecture

The pacman game is implemented using multithreaded programing with the GO lenguage, because this aloud us to run the ghost enemies and the pacman in a independent way, we use the following components to reach the solution:

## Pacman:

This element is for the player, with this element we track the position of pacman and it is loaded into the maze at the same time the maze is loaded; we handle its movements with the arrow keys presses with a method in the loop of the game.
Pacman handle 3 diferent posibilities when it moves, it can eat a dot in the game, eat a pill and aloud to eat the ghost, and to hit wall, that is works as a movement cancelled.
	
Valid keyboard inputs to move the pacman: ← , ↑ , → , ↓
	
The pacman has 3 lives that in the game we relate this to a var *lives* with a  as a initial value, the game loop will repeat if the player has at least 1 lives, onces the user loses all their lives, the game will end. 

## Ghost enemies

The enemies elements works each one in a thread and we generation the movement of the ghost with a random number from 0 to 13, and each time the value of this number is one of the following, it will make the correspondes movement, this is a very basic way to make the ghost moves by itself:
	 
`dir : is the int value that determin the ghost direction.`
		
```
  0, 4, 8, 9, 12  ->  Up
        1, 5, 13  ->  Down
        2, 6, 10  ->  Right
        3, 7, 11  ->  Left
```
		
In the same way as pacman, the ghost handle the collision with the walls, and it acts as a cancelled movement.
The amount of ghost in the game is selected by the player before the game begin, the max value of ghost is 12.
	
```go
type Ghost struct{
  position sprite
  status GhostStatus
}
```

## Maze config

The maze for the game is loaded from the *maze01.txt* file in the repository and print it in a infinite loop line by line in the console, this file has the necesary information to build the maze in our program an handle the walls and the posible postions for the enemy and pacman.
The maze is cleaned and printed each loop to handle the correct visualization as a game when a ghost or pacman moved.

```go
var maze []string
```

**Maze Data:**
```
 # -> represents a wall
 . -> represents a dot
 P -> represents the player
 G -> represents the ghosts
 X -> represents the power up pills
```


## Pill

The pill in the pacman game is another important element because this is the one that alouds pacman to power up and be able to eat the ghost for a certain amout of time, this pills are generated as the other elements in the same moment that we print the maze in the console.
To handle the fuction of the pill was necesaryu to add some methods to do the process in the pill and to take time this power up should take, and to change the configuration of a certain value `GhostStatus` in the ghost struct, during this power up is active the way the ghost are printed in the maze will change to let the player know that the ghost are vulnerable. In the pill code we take care to handle different small cases that sometimes in the game happen, for example we need to call the method to process the pill  functionality in a asynchronous way to make sure that when we eat a second pill before the fisrt one is over, we won´t lose the extra time the second one should give us.
	
```go
var pillTimer *time.Timer
var pillMx sync.Mutex
```

## Movement


In here we have two different kind of move fuction, the `moveGhost` fuction that aloud us to move in a random way the ghost running each time of loop cycle in each of the ghost threats, and the second one is the fucntion `movePlayer` that received and input from the user and the main object is to move the pacman through the maze.
Both of fuctions to move the elements, is properly handle it with the fuction `makeMove` to take care is a valid move for the corresponding element and to handle when this hits a wall.

# Diagram 
	
	![Pacman Diagram](Images/Pacman Diagram.jpg)
	
	
## UI Configuration

The ui is done printing it in the console, we reached a friendly ui with  the following struct `config` , with that we related the different elements of the maze with a certain emoji from the *config.json*.
	
```go
	{
		"player": "😃",
		"ghost": "👻",
		"wall": "  ",
		"dot": "▫️ ",
		"pill": "💊",
		"death": "💀",
		"space": "  ",
		"use_emoji": true,
		"ghost_blue": "🥶",
		"pillTime": 10
	}
```

	
