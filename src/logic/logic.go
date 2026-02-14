package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"regexp"
)

type JsonConfig struct {
	Platform  string`json:"platform"`
	Game_path string`json:"game_path"`
	Extension string`json:"extension"`
	Bin				string`json:"bin"`
	Args 		[]string`json:"args"`
	Regex 		string`json:"regex"` 
}

type Config struct {
	PlatformG 	[]string
	Game_pathG 	[]string
	ExtensionG 	[]string
	BinG 				[]string
	ArgsG 		[][]string
	RegexG     	[]string
}

type GameBuffer struct {
	GamePlatform  string 
	GamePath 			string // game_path + item + extension
	GameBin 			string
	GameArgs    []string
	GameName 			string // regex of path.
}

type GameProcess struct {
	Cmd *exec.Cmd
	Buffer *bytes.Buffer
}

func ScanDir(buffer *[]string, path string, suffix string) {

	if len(path) <= 0 {log.Fatal("Path is empty!")}	
	
	// we do this because filepath.Abs does not append / at the end.
	// because of this when we scan into a sub folder it will 
	// try to read the path and item with out a / to sperate them
	// path = /home/user/Videos
	// item = funny.mp4
	// output: /home/user/Videosfunny.mp4 
	if path[len(path)-1] != '/' {path = path+"/"}
	
	// a better implentation is to check the entire string for ~
	// then replace anything before and ~ with homedir
	
	// but its designed for me so i can deal with the slop i have created

	var new_path string
	if path[0:1] == "~" {
		home, _ := os.UserHomeDir()
		new_path = home + path[1:]
	} else {new_path = path}
	
	entries, err := os.ReadDir(new_path)
	if err != nil {log.Fatal(err)}
	
	for _, entry := range entries {
		abs_path, _ := filepath.Abs(new_path+entry.Name())

		if entry.IsDir() {
			ScanDir(buffer, abs_path, suffix)		
		} else if abs_path[len(abs_path)-len(suffix):] == suffix {
			//fmt.Printf("%s\n", abs_path)
			*buffer = append(*buffer, abs_path)
		}
	}
}

func RunGame(bin string, args []string) (*GameProcess, error) {
	cmd := exec.Command(bin, args...)
	
	var gameBuf bytes.Buffer
	cmd.Stdout = &gameBuf
	cmd.Stderr = &gameBuf

	err := cmd.Start()
	if err != nil {
		return nil, err	
	}
	return &GameProcess{Cmd: cmd, Buffer: &gameBuf}, nil
}

func ShouldGameDie(cmd *GameProcess) bool {
	if cmd == nil || cmd.Cmd == nil || cmd.Cmd.Process == nil {return false}

	var wait_status syscall.WaitStatus
	var rusage syscall.Rusage

	pid, err := syscall.Wait4(cmd.Cmd.Process.Pid, &wait_status, syscall.WNOHANG, &rusage)

	if err == nil && pid == cmd.Cmd.Process.Pid {
		fmt.Printf("Game Died: %s\n", cmd.Cmd.String())
		return true
	}

	return false
}

// we need to create a custom buffer for this
func ReadJsonConfigs(bufferConfig *Config) {
	var buffer []string
	ScanDir(&buffer, "./config", "json")

	for _, config_file := range buffer {
	  file, err := os.Open(config_file)
	  if err != nil {log.Fatal(err)}
	  defer file.Close()

	  var config JsonConfig
	  read := json.NewDecoder(file)
	  err = read.Decode(&config)
	  if err != nil {log.Fatal(err)}

		bufferConfig.PlatformG = append(bufferConfig.PlatformG, config.Platform)
		bufferConfig.Game_pathG= append(bufferConfig.Game_pathG, config.Game_path)
		bufferConfig.ExtensionG = append(bufferConfig.ExtensionG, config.Extension)
		bufferConfig.BinG = append(bufferConfig.BinG, config.Bin)
		bufferConfig.ArgsG = append(bufferConfig.ArgsG, config.Args)
		bufferConfig.RegexG = append(bufferConfig.RegexG, config.Regex)
	}
}

func RegexName(path string, regex string) string {
	re, err := regexp.Compile(regex)
	if err != nil {return path}

	matches := re.FindStringSubmatch(path)

	if len(matches) > 1 {
		return matches[1]
	}

	return path

}
