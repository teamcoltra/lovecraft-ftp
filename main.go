// Package main implements a simple virtual FTP server with a fake file system.
// It supports basic FTP commands such as USER, PASS, PWD, CWD, LIST, RETR, PASV, PORT, and QUIT.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Configuration constants
const (
	// listenAddress is the TCP address on which the FTP server listens.
	listenAddress = ":21"
	// welcomeMessage is sent to the client upon connection.
	welcomeMessage = "Welcome to the file server, if you are not authorized please disconnect. For support please email lovecraftftp@gmail.com"
	// resumeText is the dummy content returned for file downloads.
	resumeText = `Hey there,
As you might have guessed this file doesn't exist. However, what does exist is my desire to get a job. If you're looking for a Go developer, PHP developer, or a sysadmin type role please email me at teamcoltra@gmail.com. I really love creative problem solving, web scraping, and in general working on cool projects.

If you don't want to hire me you can always drop a star at https://github.com/teamcoltra/lovecraft-ftp .

Thanks for your time,
Travis Peacock
`
	// pasvIP should be set to the IP address that clients can reach.
	pasvIP = "127.0.0.1"
)

//
// Virtual File System Structures
//

// FSNode represents a file or directory node in the virtual file system.
type FSNode struct {
	Name     string    // Name of the file or directory.
	IsDir    bool      // Is true if the node is a directory.
	Children []*FSNode // Children nodes; valid only if IsDir is true.
	Size     int64     // Fake file size in bytes.
}

// FindChild returns the child node with the given name, or nil if not found.
func (node *FSNode) FindChild(childName string) *FSNode {
	for _, child := range node.Children {
		if child.Name == childName {
			return child
		}
	}
	return nil
}

//
// Virtual File System Generation
//

// createFileSystem builds the virtual file system structure and returns the root node.
func createFileSystem() *FSNode {
	root := &FSNode{Name: "/", IsDir: true}

	// Top-level directories.
	docsDir := &FSNode{Name: "documents", IsDir: true}
	picsDir := &FSNode{Name: "pictures", IsDir: true}
	downloadsDir := &FSNode{Name: "downloads", IsDir: true}
	applicationsDir := &FSNode{Name: "applications", IsDir: true}

	// documents subdirectories.
	passwordsDir := &FSNode{Name: "passwords", IsDir: true}
	backupsDir := &FSNode{Name: "backups", IsDir: true}
	recordsDir := &FSNode{Name: "records", IsDir: true}
	// records subdirectories.
	bankRecords := &FSNode{Name: "bank", IsDir: true}
	bitcoinRecords := &FSNode{Name: "bitcoin", IsDir: true}
	recordsDir.Children = append(recordsDir.Children, bankRecords, bitcoinRecords)
	// Additional random directories in documents.
	harmonyDir := &FSNode{Name: "harmony", IsDir: true}
	echoDir := &FSNode{Name: "echo", IsDir: true}
	legacyDir := &FSNode{Name: "legacy", IsDir: true}

	docsDir.Children = append(docsDir.Children, passwordsDir, backupsDir, recordsDir, harmonyDir, echoDir, legacyDir)

	// pictures subdirectories.
	pornDir := &FSNode{Name: "porn", IsDir: true}
	weddingDir := &FSNode{Name: "wedding_2023", IsDir: true}
	secretDir := &FSNode{Name: "secret", IsDir: true}
	picsDir.Children = append(picsDir.Children, pornDir, weddingDir, secretDir)

	// downloads subdirectories.
	usenetDir := &FSNode{Name: "Usenet", IsDir: true}
	torrentsDir := &FSNode{Name: "Torrents", IsDir: true}
	downloadsDir.Children = append(downloadsDir.Children, usenetDir, torrentsDir)

	// applications subdirectories.
	gamesDir := &FSNode{Name: "games", IsDir: true}
	applicationsDir.Children = append(applicationsDir.Children, gamesDir)

	root.Children = append(root.Children, docsDir, picsDir, downloadsDir, applicationsDir)

	// Generate files in every directory recursively.
	generateFilesRecursively(root, "")
	return root
}

// generatePornTitle returns a randomly generated porn-themed title.
func generatePornTitle() string {
	innuendos := []string{
		"SlipperyWhenWet", "FullThrottle", "ComeHither", "DeepDesires", "RacySecrets",
		"VelvetTouch", "HotNReady", "WildAffair", "ForbiddenFruit", "SlickOperator",
		"SensualSizzle", "BareAll", "AfterDark", "SecretPassion", "SteamyNights",
		"VelvetWhisper", "LustfulWhimsy", "TantalizingTease", "SmolderingHeat", "ProvocativePulse",
	}

	sexualPreferences := []string{"Straight", "Gay", "Bi", "Lesbian", "Trans"}
	adjectives := []string{
		"Naughty", "Steamy", "Hot", "Fiery", "Wild",
		"Racy", "Sultry", "Sensual", "Lusty", "Spicy",
		"Scandalous", "Untamed", "Ravishing", "Tempting", "Seductive",
		"Passionate", "Erotic", "Feverish", "Racy", "Alluring",
	}
	jobTitles := []string{
		"Nurse", "Librarian", "Teacher", "Mechanic", "Secretary",
		"PoliceOfficer", "Chef", "Firefighter", "Athlete", "Accountant",
		"Engineer", "Bartender", "Lawyer", "Pilot", "Paramedic",
		"Receptionist", "Designer", "Journalist", "Photographer", "Dentist",
	}
	dickWords := []string{
		"Knob", "Dick", "Johnson", "Shaft", "Cock",
		"Rod", "Member", "Tool", "Pecker", "Prick",
		"Manhood", "Wand", "Junk", "Twig", "Beaver",
		"Monster", "Staff", "Sausage", "Cucumber", "Spur",
	}
	dickAdjectives := []string{
		"Hard", "Throbbing", "Stiff", "RockHard", "Pulsating",
		"Massive", "Vigorous", "Mighty", "Raging", "Dominant",
		"Bulging", "Strapping", "Fierce", "Robust", "Lusty",
		"Potent", "Vigorous", "Electric", "Savage", "Fevered",
	}
	menFirstNames := []string{
		"Mike", "John", "Bob", "Steve", "Tony",
		"Dave", "Mark", "Luke", "Jake", "Ryan",
		"Alex", "Chris", "Nick", "Sam", "Brian",
		"Dan", "Tom", "Greg", "Eric", "Paul",
	}
	womenFirstNames := []string{
		"Candy", "Destiny", "Sugar", "Roxy", "Tiffany",
		"Angel", "Cherry", "Ruby", "Kitten", "Daisy",
		"Bonnie", "Lola", "Lacey", "Ginger", "Penny",
		"Skye", "Jasmine", "Bella", "Misty", "Summer",
	}
	womenLastNames := []string{
		"Devine", "Lovin", "Sweet", "Delight", "Satin",
		"Passion", "Bliss", "Heaven", "Desire", "Charm",
		"Fever", "Luscious", "Velvet", "Sin", "Dream",
		"Sparkle", "Rose", "Star", "Mystique", "Flame",
	}

	// Helper functions to generate random names.
	generateManName := func() string { return menFirstNames[rand.Intn(len(menFirstNames))] }
	generateManPornLastName := func() string {
		return dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))]
	}
	generateWomanName := func() string { return womenFirstNames[rand.Intn(len(womenFirstNames))] }
	generateWomanLastName := func() string { return womenLastNames[rand.Intn(len(womenLastNames))] }

	option1 := innuendos[rand.Intn(len(innuendos))] + " " +
		sexualPreferences[rand.Intn(len(sexualPreferences))] + " " +
		adjectives[rand.Intn(len(adjectives))] + " " +
		jobTitles[rand.Intn(len(jobTitles))]

	option2 := "Fake-" + jobTitles[rand.Intn(len(jobTitles))] + "-" +
		generateWomanName() + "-" + generateWomanLastName()

	option3 := "Fake-" + jobTitles[rand.Intn(len(jobTitles))] + "-" +
		generateManName() + "-" + generateManPornLastName()

	option4 := adjectives[rand.Intn(len(adjectives))] + " " +
		jobTitles[rand.Intn(len(jobTitles))] + " " + generateWomanName()

	option5 := innuendos[rand.Intn(len(innuendos))] + " " +
		dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))]

	option6 := jobTitles[rand.Intn(len(jobTitles))] + " of " +
		generateWomanName() + " " + generateWomanLastName()

	option7 := innuendos[rand.Intn(len(innuendos))] + "-Fake-" +
		generateManName() + "-" + generateManPornLastName()

	option8 := adjectives[rand.Intn(len(adjectives))] + " " +
		sexualPreferences[rand.Intn(len(sexualPreferences))] + " " +
		jobTitles[rand.Intn(len(jobTitles))] + " & " +
		generateWomanName() + "'s " + generateWomanLastName()

	option9 := "XXX " + option1

	option10 := jobTitles[rand.Intn(len(jobTitles))] + "-" +
		adjectives[rand.Intn(len(adjectives))] + "-" +
		innuendos[rand.Intn(len(innuendos))]

	option11 := "Fake-" + generateWomanName() + "-" + generateWomanLastName() + "-" + jobTitles[rand.Intn(len(jobTitles))]

	option12 := dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))] + " meets " + generateWomanName()

	option13 := innuendos[rand.Intn(len(innuendos))] + " & " + generateManName() + "'s " +
		dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))]

	option14 := sexualPreferences[rand.Intn(len(sexualPreferences))] + " " +
		jobTitles[rand.Intn(len(jobTitles))] + " " +
		dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))]

	option15 := adjectives[rand.Intn(len(adjectives))] + " " +
		generateWomanName() + " " + jobTitles[rand.Intn(len(jobTitles))] + " Fantasy"

	option16 := generateManName() + " and the " +
		dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))]

	option17 := innuendos[rand.Intn(len(innuendos))] + " " +
		jobTitles[rand.Intn(len(jobTitles))] + " featuring " +
		generateWomanName() + "'s " + generateWomanLastName()

	option18 := "Fake-" + generateManName() + "-" + dickAdjectives[rand.Intn(len(dickAdjectives))] +
		dickWords[rand.Intn(len(dickWords))] + " vs Fake-" +
		generateWomanName() + "-" + generateWomanLastName()

	option19 := adjectives[rand.Intn(len(adjectives))] + " " +
		sexualPreferences[rand.Intn(len(sexualPreferences))] + " " +
		innuendos[rand.Intn(len(innuendos))] + " " +
		jobTitles[rand.Intn(len(jobTitles))]

	option20 := jobTitles[rand.Intn(len(jobTitles))] + " X " + generateManName() + "'s " +
		dickAdjectives[rand.Intn(len(dickAdjectives))] + dickWords[rand.Intn(len(dickWords))]

	options := []string{
		option1, option2, option3, option4, option5,
		option6, option7, option8, option9, option10,
		option11, option12, option13, option14, option15,
		option16, option17, option18, option19, option20,
	}

	title := options[rand.Intn(len(options))]
	roll := rand.Intn(100)
	if roll < 10 {
		title = "xxx " + title
	} else if roll >= 90 {
		title = title + " xxx"
	}
	return title
}

// generatePornFilename returns a slugified filename with a random video extension.
func generatePornFilename() string {
	videoExtensions := []string{".mp4", ".avi", ".mkv", ".flv", ".wmv"}
	slug := convertToSlug(generatePornTitle())
	ext := videoExtensions[rand.Intn(len(videoExtensions))]
	return slug + ext
}

// generateFileName returns a plausible filename based on the provided file category.
func generateFileName(category string) string {
	// General adjectives.
	adjectives := []string{
		"quick", "happy", "bright", "silent", "mellow",
		"brisk", "calm", "clever", "daring", "elegant",
		"fancy", "gentle", "jolly", "lively", "polite",
		"quiet", "rapid", "shiny", "smiling", "witty",
	}

	// Nouns and extensions for different categories.
	pictureNouns := []string{
		"sunset", "mountain", "beach", "forest", "cityscape",
		"portrait", "landscape", "snapshot", "selfie", "reflection",
		"vista", "waterfall", "garden", "skyline", "horizon",
	}
	pictureExts := []string{".jpg", ".png", ".gif", ".bmp"}

	documentNouns := []string{
		"report", "proposal", "memo", "summary", "draft",
		"invoice", "agenda", "minutes", "letter", "notes",
		"analysis", "blueprint", "plan", "overview", "abstract",
		"manual", "document", "research", "file", "paper",
	}
	documentExts := []string{".doc", ".pdf", ".txt", ".rtf", ".odt"}

	downloadNouns := []string{
		"installer", "update", "package", "archive", "setup",
		"bundle", "release", "version", "patch", "module",
		"download", "resource", "addon", "toolkit", "driver",
		"script", "library", "binary", "compiler", "framework",
	}
	downloadExts := []string{".zip", ".rar", ".exe", ".msi", ".tar.gz"}

	applicationNouns := []string{
		"calculator", "editor", "notepad", "browser", "player",
		"manager", "tracker", "organizer", "viewer", "mailer",
		"converter", "explorer", "designer", "scheduler", "recorder",
		"terminal", "dashboard", "monitor", "assistant", "studio",
	}
	applicationExts := []string{".app", ".exe", ".bin"}

	gameNouns := []string{
		"adventure", "quest", "battle", "arena", "saga",
		"challenge", "odyssey", "mission", "struggle", "duel",
		"clash", "legend", "racer", "fighter", "hero",
		"escape", "survival", "chronicle", "empire", "fantasy",
	}
	gameExts := []string{".game", ".bin", ".rom", ".iso", ""}

	// Helper function to select a random element.
	randChoice := func(choices []string) string {
		return choices[rand.Intn(len(choices))]
	}

	suffix := rand.Intn(90) + 10 // a number between 10 and 99

	var noun, ext string
	switch category {
	case "pictures":
		noun = randChoice(pictureNouns)
		ext = randChoice(pictureExts)
	case "documents":
		noun = randChoice(documentNouns)
		ext = randChoice(documentExts)
	case "downloads":
		noun = randChoice(downloadNouns)
		ext = randChoice(downloadExts)
	case "applications":
		noun = randChoice(applicationNouns)
		ext = randChoice(applicationExts)
	case "game names":
		noun = randChoice(gameNouns)
		ext = randChoice(gameExts)
	default:
		noun = randChoice(documentNouns)
		ext = randChoice(documentExts)
	}

	adj := randChoice(adjectives)
	return adj + "_" + noun + "_" + fmt.Sprintf("%d", suffix) + ext
}

// convertToSlug converts a given title into a URL-friendly slug.
func convertToSlug(title string) string {
	slug := strings.ToLower(title)
	invalidCharRegexp := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = invalidCharRegexp.ReplaceAllString(slug, "")
	slug = strings.ReplaceAll(slug, " ", "-")
	multipleHyphenRegexp := regexp.MustCompile(`-+`)
	slug = multipleHyphenRegexp.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// determineCategory returns the file category based on the top-level directory of the given path.
func determineCategory(currentPath string) string {
	parts := strings.Split(filepath.Clean(currentPath), string(filepath.Separator))
	if len(parts) > 1 {
		topDir := parts[1]
		if topDir == "documents" || topDir == "pictures" || topDir == "downloads" || topDir == "applications" {
			return topDir
		}
	}
	return "documents"
}

// randomFileSize returns a randomly chosen file size from a set of predetermined sizes.
func randomFileSize() int64 {
	possibleSizes := []int64{
		69,           // repeated "69"
		6969,         // "69" twice
		69696969,     // "69" repeated four times
		420,          // "420"
		420420,       // "420" twice
		420420420420, // "420" repeated four times
		42069,        // mixed
		69420,        // mixed
		6942069,      // mixed
	}
	return possibleSizes[rand.Intn(len(possibleSizes))]
}

// generateFilesRecursively populates a directory FSNode with a random number of file nodes
// and recursively processes its subdirectories.
func generateFilesRecursively(node *FSNode, currentPath string) {
	if node.IsDir {
		// Update the current path for subdirectories.
		if node.Name != "/" {
			if currentPath == "" {
				currentPath = "/" + node.Name
			} else {
				currentPath = path.Join(currentPath, node.Name)
			}
		}

		// Determine file category based on the current path.
		category := determineCategory(currentPath)
		numFiles := rand.Intn(21) + 10 // between 10 and 30 files

		for i := 0; i < numFiles; i++ {
			var fileName string
			// For directories named "porn" (case-insensitive), use the porn filename generator.
			if strings.ToLower(node.Name) == "porn" {
				fileName = generatePornFilename()
			} else {
				fileName = generateFileName(category)
			}
			fileNode := &FSNode{
				Name:  fileName,
				IsDir: false,
				Size:  randomFileSize(),
			}
			node.Children = append(node.Children, fileNode)
		}

		// Recursively process child directories.
		for _, child := range node.Children {
			if child.IsDir {
				generateFilesRecursively(child, currentPath)
			}
		}
	}
}

// traverseFileSystem returns the FSNode corresponding to the given Unix-style path.
// It returns nil if the path does not exist in the virtual file system.
func traverseFileSystem(pathStr string) *FSNode {
	if pathStr == "/" || pathStr == "" {
		return fsRoot
	}
	parts := strings.Split(filepath.Clean(pathStr), string(filepath.Separator))
	currentNode := fsRoot
	for _, part := range parts {
		if part == "" {
			continue
		}
		currentNode = currentNode.FindChild(part)
		if currentNode == nil {
			return nil
		}
	}
	return currentNode
}

//
// Command Logging
//

// CommandLog represents a single command log entry.
type CommandLog struct {
	Timestamp string `json:"timestamp"`
	IP        string `json:"ip"`
	Command   string `json:"command"`
	Argument  string `json:"argument,omitempty"`
	CWD       string `json:"cwd"`
}

var (
	// cmdLogFile is the file where command logs are stored.
	cmdLogFile *os.File
	// logMutex protects access to cmdLogFile.
	logMutex sync.Mutex
)

// initCommandLogger initializes the command logger by opening (or creating) the log file.
func initCommandLogger() {
	var err error
	cmdLogFile, err = os.OpenFile("commands.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening command log file: %v", err)
	}
}

// logCommand writes a command log entry in JSON lines format.
func logCommand(ip, command, argument, cwd string) {
	entry := CommandLog{
		Timestamp: time.Now().Format(time.RFC3339),
		IP:        ip,
		Command:   command,
		Argument:  argument,
		CWD:       cwd,
	}
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshaling command log: %v", err)
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()
	cmdLogFile.WriteString(string(entryJSON) + "\n")
}

//
// FTP Session Handling
//

// ftpSession represents a client session for the FTP server.
type ftpSession struct {
	conn              net.Conn      // Control connection.
	reader            *bufio.Reader // Buffered reader for the control connection.
	writer            *bufio.Writer // Buffered writer for the control connection.
	cwd               string        // Current working directory.
	logPrefix         string        // Prefix used for logging messages.
	pasvListener      net.Listener  // Listener for passive mode data connection.
	activeDataAddress string        // Address for active mode data connection.
	dataConnection    net.Conn      // Established data connection.
}

// newFTPSession creates a new ftpSession for the given connection.
func newFTPSession(conn net.Conn) *ftpSession {
	// Enable TCP keep-alives.
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	return &ftpSession{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		cwd:       "/",
		logPrefix: fmt.Sprintf("[%s]", conn.RemoteAddr().String()),
	}
}

// writeLine writes a response line to the client connection.
func (s *ftpSession) writeLine(line string) error {
	_, err := s.writer.WriteString(line + "\r\n")
	if err != nil {
		return err
	}
	return s.writer.Flush()
}

// closeDataConnection closes the active data connection and any passive listener.
func (s *ftpSession) closeDataConnection() {
	if s.dataConnection != nil {
		s.dataConnection.Close()
		s.dataConnection = nil
	}
	if s.pasvListener != nil {
		s.pasvListener.Close()
		s.pasvListener = nil
	}
	s.activeDataAddress = ""
}

// getDataConnection returns a data connection based on the current session mode (passive or active).
func (s *ftpSession) getDataConnection() (net.Conn, error) {
	if s.pasvListener != nil {
		conn, err := s.pasvListener.Accept()
		if err != nil {
			return nil, err
		}
		s.dataConnection = conn
		s.pasvListener.Close()
		s.pasvListener = nil
		return conn, nil
	}
	if s.activeDataAddress != "" {
		conn, err := net.Dial("tcp", s.activeDataAddress)
		if err != nil {
			return nil, err
		}
		s.dataConnection = conn
		s.activeDataAddress = ""
		return conn, nil
	}
	return nil, fmt.Errorf("425 Use PASV or PORT/EPRT first")
}

// handleSession processes FTP commands from the client and handles file transfers.
func (s *ftpSession) handleSession() {
	defer s.conn.Close()
	log.Printf("%s New connection", s.logPrefix)
	s.writeLine("220 " + welcomeMessage)

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			log.Printf("%s Connection error: %v", s.logPrefix, err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		log.Printf("%s Received: %s", s.logPrefix, line)
		parts := strings.SplitN(line, " ", 2)
		command := strings.ToUpper(parts[0])
		argument := ""
		if len(parts) > 1 {
			argument = parts[1]
		}

		// Log the command.
		logCommand(s.conn.RemoteAddr().String(), command, argument, s.cwd)

		switch command {
		case "USER":
			log.Printf("%s Login attempt: USER %s", s.logPrefix, argument)
			s.writeLine("331 Username OK, need password.")
		case "PASS":
			log.Printf("%s User logged in", s.logPrefix)
			s.writeLine("230 Login successful.")
		case "SYST":
			s.writeLine("215 UNIX Type: L8")
		case "PWD":
			s.writeLine(fmt.Sprintf(`257 "%s" is the current directory.`, s.cwd))
		case "TYPE":
			if strings.ToUpper(argument) == "I" {
				s.writeLine("200 Switching to Binary mode.")
			} else {
				s.writeLine("200 OK")
			}
		case "CWD":
			var newPath string
			if strings.HasPrefix(argument, "/") {
				newPath = argument
			} else {
				newPath = path.Join(s.cwd, argument)
			}
			if node := traverseFileSystem(newPath); node != nil && node.IsDir {
				s.cwd = path.Clean(newPath)
				log.Printf("%s Changed directory to %s", s.logPrefix, s.cwd)
				s.writeLine("250 Directory successfully changed.")
			} else {
				s.writeLine("550 Failed to change directory.")
			}
		case "PASV":
			s.closeDataConnection()
			listener, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				s.writeLine("425 Can't open passive connection.")
				break
			}
			s.pasvListener = listener
			addr := listener.Addr().(*net.TCPAddr)
			ipParts := strings.Split(pasvIP, ".")
			p1 := addr.Port / 256
			p2 := addr.Port % 256
			response := fmt.Sprintf("227 Entering Passive Mode (%s,%s,%s,%s,%d,%d).",
				ipParts[0], ipParts[1], ipParts[2], ipParts[3], p1, p2)
			s.writeLine(response)
		case "EPSV":
			s.closeDataConnection()
			listener, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				s.writeLine("425 Can't open passive connection.")
				break
			}
			s.pasvListener = listener
			addr := listener.Addr().(*net.TCPAddr)
			response := fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|)", addr.Port)
			s.writeLine(response)
		case "PORT":
			parts := strings.Split(argument, ",")
			if len(parts) != 6 {
				s.writeLine("501 Syntax error in parameters or arguments.")
				break
			}
			ipAddr := strings.Join(parts[0:4], ".")
			p1, err1 := strconv.Atoi(parts[4])
			p2, err2 := strconv.Atoi(parts[5])
			if err1 != nil || err2 != nil {
				s.writeLine("501 Syntax error in parameters or arguments.")
				break
			}
			port := p1*256 + p2
			s.activeDataAddress = fmt.Sprintf("%s:%d", ipAddr, port)
			s.writeLine("200 PORT command successful.")
		case "EPRT":
			delimiter := string(argument[0])
			fields := strings.Split(argument, delimiter)
			if len(fields) < 4 {
				s.writeLine("501 Syntax error in parameters or arguments.")
				break
			}
			ipAddr := fields[2]
			port, err := strconv.Atoi(fields[3])
			if err != nil {
				s.writeLine("501 Syntax error in parameters or arguments.")
				break
			}
			s.activeDataAddress = fmt.Sprintf("%s:%d", ipAddr, port)
			s.writeLine("200 EPRT command successful.")
		case "LIST":
			conn, err := s.getDataConnection()
			if err != nil {
				s.writeLine("425 " + err.Error())
				break
			}
			s.writeLine("150 Opening data connection for directory list.")
			node := traverseFileSystem(s.cwd)
			if node == nil || !node.IsDir {
				s.writeLine("550 Not a directory.")
				s.closeDataConnection()
				break
			}
			var listing bytes.Buffer
			for _, child := range node.Children {
				if child.IsDir {
					listing.WriteString(fmt.Sprintf("drwxr-xr-x 1 ftp ftp %12d Jan 01 00:00 %s\r\n", 0, child.Name))
				} else {
					listing.WriteString(fmt.Sprintf("-rw-r--r-- 1 ftp ftp %12d Jan 01 00:00 %s\r\n", child.Size, child.Name))
				}
			}
			conn.Write(listing.Bytes())
			conn.Close()
			s.closeDataConnection()
			s.writeLine("226 Directory send OK.")
		case "RETR":
			targetPath := path.Join(s.cwd, argument)
			node := traverseFileSystem(targetPath)
			if node == nil || node.IsDir {
				log.Printf("%s RETR failed. Path %s not found.", s.logPrefix, targetPath)
				s.writeLine("550 File not found.")
				break
			}
			conn, err := s.getDataConnection()
			if err != nil {
				s.writeLine("425 " + err.Error())
				break
			}
			s.writeLine("150 Opening data connection for file transfer.")
			// In this demo, the file contents are simulated.
			conn.Write([]byte(resumeText))
			conn.Close()
			s.closeDataConnection()
			s.writeLine("226 Transfer complete.")
		case "QUIT":
			s.writeLine("221 Goodbye.")
			log.Printf("%s Connection closed by client.", s.logPrefix)
			return
		default:
			s.writeLine("502 Command not implemented.")
		}
	}
}

//
// Main entry point
//

var fsRoot *FSNode

// main initializes the command logger, creates the virtual file system, and starts the FTP server.
func main() {
	initCommandLogger()
	defer cmdLogFile.Close()
	fsRoot = createFileSystem()
	log.Printf("Starting virtual FTP server on %s", listenAddress)
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		session := newFTPSession(conn)
		go session.handleSession()
	}
}
