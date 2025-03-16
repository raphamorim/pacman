package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Game dimensions
	width  = 30
	height = 20

	// Game characters
	pacman      = '❤'
	ghost       = '⚉'
	dot         = '·'
	wall        = '█'
	powerPellet = '●'
	emptySpace  = ' '

	// Game states
	playing = iota
	won
	lost
	paused
)

// Direction represents movement direction
type Direction int

const (
	up Direction = iota
	down
	left
	right
)

// Ghost represents an enemy
type Ghost struct {
	x, y            int
	direction       Direction
	frightened      bool
	frightenedTimer int
}

// Model represents the game state
type Model struct {
	pacmanX, pacmanY int
	pacmanDirection  Direction
	ghosts           []Ghost
	score            int
	lives            int
	board            [][]rune
	gameState        int
	timer            timer.Model
	width            int
	height           int
	dotCount         int
	powerMode        bool
	powerTimer       int
	level            int
}

var (
	// Styles
	pacmanStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))            // Yellow
	ghostStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#d7005f"))            // Red
	frightenedGhostStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))            // Blue
	wallStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff"))            // Fushia
	dotStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))            // White
	powerPelletStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))            // White
	scoreStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))            // Green
	gameOverStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true) // Red, bold
	winStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true) // Green, bold
)

// Initial maze layout - W is a wall, D is a dot, P is a power pellet, space is empty
var initialMaze = []string{
	"WWWWWWWWWWWWWWWWWWWWWWWWWWWWWW",
	"WDDDDDDDDDDDDWWDDDDDDDDDDDDDDW",
	"WDWWWWDWWWWWDWWDWWWWWDWWWWWDWW",
	"WPDWWWDWWWWWDWWDWWWWWDWWWWWDPW",
	"WDWWWWDWWWWWDWWDWWWWWDWWWWWDWW",
	"WDDDDDDDDDDDDDDDDDDDDDDDDDDDDW",
	"WDWWWWDWWDWWWWWWWWDWWDWWWWWDWW",
	"WDWWWWDWWDWWWWWWWWDWWDWWWWWDWW",
	"WDDDDDDWWDDDDWWDDDDWWDDDDDDWW",
	"WWWWWWDWWWWW WWWW WWWWDWWWWWW",
	"     WDWWWWW WWWW WWWWDW     ",
	"     WDWW          WWDW     ",
	"     WDWW WWWWWWWW WWDW     ",
	"WWWWWWDWW W      W WWDWWWWWWW",
	"      D   W      W   D      ",
	"WWWWWWDWW WWWWWWWW WWDWWWWWWW",
	"     WDWW          WWDW     ",
	"     WDWW WWWWWWWW WWDW     ",
	"WWWWWWDWW WWWWWWWW WWDWWWWWWW",
	"WDDDDDDDDDDDDWWDDDDDDDDDDDDWW",
	"WDWWWWDWWWWWDWWDWWWWWDWWWWWDWW",
	"WPDWWWDWWWWWDWWDWWWWWDWWWWWDPW",
	"WDWWWWDWWWWWDWWDWWWWWDWWWWWDWW",
	"WDDDDDDWWDDDDDDDDDDWWDDDDDDWW",
	"WDWWWWDWWDWWWWWWWWDWWDWWWWWDWW",
	"WDWWWWDWWDWWWWWWWWDWWDWWWWWDWW",
	"WDDDDDDDDDDDDWWDDDDDDDDDDDDWW",
	"WWWWWWWWWWWWWWWWWWWWWWWWWWWWWW",
}

// Initialize a new game
func initialModel() Model {
	// Create the game board
	board := make([][]rune, height)
	dotCount := 0

	for y := 0; y < height; y++ {
		board[y] = make([]rune, width)
		if y < len(initialMaze) {
			row := initialMaze[y]
			for x := 0; x < width && x < len(row); x++ {
				switch row[x] {
				case 'W':
					board[y][x] = wall
				case 'D':
					board[y][x] = dot
					dotCount++
				case 'P':
					board[y][x] = powerPellet
				case ' ':
					board[y][x] = emptySpace
				default:
					board[y][x] = emptySpace
				}
			}
		}
	}

	// Create ghosts
	ghosts := []Ghost{
		{x: 14, y: 10, direction: up},
		{x: 15, y: 10, direction: up},
		{x: 14, y: 11, direction: up},
		{x: 15, y: 11, direction: up},
	}

	return Model{
		pacmanX:         1,
		pacmanY:         1,
		pacmanDirection: right,
		ghosts:          ghosts,
		score:           0,
		lives:           3,
		board:           board,
		gameState:       playing,
		timer:           timer.NewWithInterval(time.Second/5, time.Millisecond*120),
		width:           width,
		height:          height,
		dotCount:        dotCount,
		powerMode:       false,
		powerTimer:      0,
		level:           1,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.timer.Init(), tea.EnterAltScreen)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.pacmanDirection = up
		case "down", "j":
			m.pacmanDirection = down
		case "left", "h":
			m.pacmanDirection = left
		case "right", "l":
			m.pacmanDirection = right
		case "p":
			// Toggle pause
			if m.gameState == paused {
				m.gameState = playing
			} else if m.gameState == playing {
				m.gameState = paused
			}
		case "r":
			// Restart game
			return initialModel(), nil
		}

	case timer.TickMsg:
		if m.gameState == playing {
			// Move Pacman
			newX, newY := m.pacmanX, m.pacmanY
			switch m.pacmanDirection {
			case up:
				newY--
			case down:
				newY++
			case left:
				newX--
			case right:
				newX++
			}

			// Check for wrap-around
			if newX < 0 {
				newX = m.width - 1
			} else if newX >= m.width {
				newX = 0
			}
			if newY < 0 {
				newY = m.height - 1
			} else if newY >= m.height {
				newY = 0
			}

			// Check for collisions with walls
			if newX >= 0 && newX < m.width && newY >= 0 && newY < m.height && m.board[newY][newX] != wall {
				m.pacmanX, m.pacmanY = newX, newY

				// Check for dots
				if m.board[m.pacmanY][m.pacmanX] == dot {
					m.board[m.pacmanY][m.pacmanX] = emptySpace
					m.score += 10
					m.dotCount--
				} else if m.board[m.pacmanY][m.pacmanX] == powerPellet {
					m.board[m.pacmanY][m.pacmanX] = emptySpace
					m.score += 50
					m.powerMode = true
					m.powerTimer = 20 // Last for 20 ticks

					// Make ghosts frightened
					for i := range m.ghosts {
						m.ghosts[i].frightened = true
						m.ghosts[i].frightenedTimer = 20
					}
				}
			}

			// Check if all dots are eaten
			if m.dotCount <= 0 {
				m.gameState = won
			}

			// Move ghosts
			for i := range m.ghosts {
				// Decrease frightened timer if active
				if m.ghosts[i].frightened {
					m.ghosts[i].frightenedTimer--
					if m.ghosts[i].frightenedTimer <= 0 {
						m.ghosts[i].frightened = false
					}
				}

				// Simple ghost AI
				if rand.Intn(100) < 20 { // 20% chance to change direction
					m.ghosts[i].direction = Direction(rand.Intn(4))
				}

				// Move ghost
				newX, newY := m.ghosts[i].x, m.ghosts[i].y
				switch m.ghosts[i].direction {
				case up:
					newY--
				case down:
					newY++
				case left:
					newX--
				case right:
					newX++
				}

				// Check for wrap-around
				if newX < 0 {
					newX = m.width - 1
				} else if newX >= m.width {
					newX = 0
				}
				if newY < 0 {
					newY = m.height - 1
				} else if newY >= m.height {
					newY = 0
				}

				// Check for collisions with walls
				if newX >= 0 && newX < m.width && newY >= 0 && newY < m.height && m.board[newY][newX] != wall {
					m.ghosts[i].x, m.ghosts[i].y = newX, newY
				} else {
					// Ghost hit a wall, change direction
					m.ghosts[i].direction = Direction(rand.Intn(4))
				}

				// Check for collision with Pacman
				if m.ghosts[i].x == m.pacmanX && m.ghosts[i].y == m.pacmanY {
					if m.ghosts[i].frightened {
						// Pacman eats ghost
						m.score += 200
						m.ghosts[i].x = 14 + (i % 2)
						m.ghosts[i].y = 10 + (i / 2)
						m.ghosts[i].frightened = false
					} else {
						// Ghost catches Pacman
						m.lives--
						if m.lives <= 0 {
							m.gameState = lost
						} else {
							// Reset Pacman position
							m.pacmanX = 1
							m.pacmanY = 1
						}
					}
				}
			}

			// Update power mode timer
			if m.powerMode {
				m.powerTimer--
				if m.powerTimer <= 0 {
					m.powerMode = false
				}
			}
		}

		return m, m.timer.Start()
	}

	m.timer, cmd = m.timer.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	s := ""

	// Game area
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			// Check if there's a ghost at this position
			isGhost := false
			isFrightened := false
			for _, g := range m.ghosts {
				if g.x == x && g.y == y {
					isGhost = true
					isFrightened = g.frightened
					break
				}
			}

			// Check if Pacman is at this position
			isPacman := m.pacmanX == x && m.pacmanY == y

			if isPacman {
				s += pacmanStyle.Render(string(pacman))
			} else if isGhost {
				if isFrightened {
					s += frightenedGhostStyle.Render(string(ghost))
				} else {
					s += ghostStyle.Render(string(ghost))
				}
			} else {
				switch m.board[y][x] {
				case wall:
					s += wallStyle.Render(string(wall))
				case dot:
					s += dotStyle.Render(string(dot))
				case powerPellet:
					s += powerPelletStyle.Render(string(powerPellet))
				default:
					s += " "
				}
			}
		}
		s += "\n"
	}

	// Game info
	scoreBar := lipgloss.NewStyle().Background(lipgloss.Color("#000000")).Foreground(lipgloss.Color("#ff00af")).Bold(true)
	s += scoreBar.Render(fmt.Sprintf("\nScore: %d  Lives: %d  Level: %d\n", m.score, m.lives, m.level))

	bottomBar := lipgloss.NewStyle().Background(lipgloss.Color("#afffff")).Foreground(lipgloss.Color("#ff00af"))

	s += "\n"

	// Game state messages
	if m.gameState == won {
		s += winStyle.Render("\nYOU WIN! Press 'r' to play again or 'q' to quit.\n")
	} else if m.gameState == lost {
		s += gameOverStyle.Render("\nGAME OVER! Press 'r' to play again or 'q' to quit.\n")
	} else if m.gameState == paused {
		s += bottomBar.Render("Game Paused. Press 'p' to continue.\n")
	} else {
		s += bottomBar.Render("Controls: arrow keys to move, 'p' to pause, 'q' to quit, 'r' to restart")
	}

	s += "\n"
	s += bottomBar.Render("(◕‿◕) Use Rio terminal with retro arch shaders for a better experience!")

	return s
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
