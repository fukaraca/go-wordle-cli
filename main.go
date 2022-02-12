package main

import (
	"bufio"
	"embed"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode"
)

import "github.com/fatih/color"

var bare = color.New(color.BgWhite, color.Bold, color.FgBlack)
var bingo = color.New(color.BgGreen, color.Bold, color.FgBlack)
var exist = color.New(color.BgYellow, color.Bold, color.FgBlack)
var sep = color.New(color.FgBlack, color.Bold)
var info = color.New(color.BgBlack, color.Bold, color.FgHiWhite)
var erroR = color.New(color.FgRed)
var notEx = color.New(color.FgRed)
var reader *bufio.Reader

func main() {

	reader = bufio.NewReader(os.Stdin)
	for {
		CallClear()
		info.Println("Simple Wordle game ")
		info.Println("---------------------")
		info.Println("Type 'start' to begin!\nType 'help' to read info \nPres CTRL+C to abort")
		erroR.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			erroR.Println(err)
		}
		text = replacer(text)

		if text == "start" {
			go startGame()
			break
		}
		if text == "help" {
			info.Println("Wordle is simply a five-letter word that you have to guess within six tries.")
			bingo.Println("If any letter matches by index it is highlighted green.")
			exist.Println("In case, letter is existed in Wordle but index is wrong, it is highlighted yellow ")
			bare.Println("Otherwise letter is highlighted white")
			info.Println("To skip press enter")
			erroR.Print("-> ")
			_, _ = reader.ReadString('\n')
		}
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

}

//game
func startGame() {
	CallClear()
	wordle, err := returnRandomWord()
	foundWordle := false
	if err != nil {
		erroR.Println("new wordle couldn't be instantiated", err)
		os.Exit(1)
	}
	matches := make(map[int32]bool)
	exists := make(map[int32]bool)
	notExists := make(map[int32]bool)
	tries := []string{}
	for {

		CallClear()
		info.Println("Let's Wordle!")
		info.Println("Type the word to submit!")
		if 6-len(tries) >= 1 {
			info.Printf("You can try %d time.\n", 6-len(tries))
		}
		sep.Print("|")
		for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			if matches[letter] {
				bingo.Printf("%s", string(letter))
			} else if exists[letter] {
				exist.Printf("%s", string(letter))
			} else if notExists[letter] {
				notEx.Printf("%s", string(letter))
			} else {
				info.Printf("%s", string(letter))
			}
			sep.Print("|")
		}
		_, _ = sep.Println("\n")

		for _, try := range tries { //previous word list
			sep.Print("|")
			for i, l := range try { //check that word letter by letter
				found := false
				if try[i] == wordle[i] {
					bingo.Printf(" %s ", string(l)) //there is match at i index
					matches[l] = true
					sep.Print("|")
					continue
				} else {
					for _, w := range wordle {
						if l == w {
							exist.Printf(" %s ", string(l)) //there is match but for another index
							exists[l] = true
							sep.Print("|")
							found = true
							break
						}
					}
				}
				if found {
					continue
				}
				//if it still couldn't get out of loop
				bare.Printf(" %s ", string(l)) //no match
				notExists[l] = true
				sep.Print("|")
			}
			sep.Print("\n")
		}
		//game over Failed
		if len(tries) == 6 && !foundWordle {
			erroR.Println("You have FAILED! Type restart or exit!")
			erroR.Print("The Wordle was:")
			bingo.Printf("%s\n", wordle)
			erroR.Print("-> ")
			text, _ := reader.ReadString('\n')
			text = replacer(text)
			if text == "restart" {
				go startGame()
				return
			} else if text == "exit" {
				erroR.Println("See you next time!")
				os.Exit(1)
			}
		}
		//game over Success
		if foundWordle {
			info.Println("Congratz!....\nType 'restart' to play again or 'exit' to leave")
			erroR.Print("-> ")
			text, _ := reader.ReadString('\n')
			text = replacer(text)
			if text == "restart" {
				go startGame()
				return
			} else if text == "exit" {
				erroR.Println("See you next time!")
				os.Exit(1)
			}
		}

		//Read next input
		erroR.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = replacer(text)

		if len(text) != 5 {
			erroR.Println("submitted word is not 5 letters at all")
			time.Sleep(2 * time.Second)
			continue
		}

		if !isItaRealWord(strings.ToLower(text)) {
			erroR.Println("it's not even a word!")
			time.Sleep(2 * time.Second)
			continue
		}
		text = strings.ToUpper(text)
		if wordle == text {
			foundWordle = true
		}
		//check if it was submitted already
		isexistTemp := false
		for _, tryT := range tries {
			if text == tryT {
				erroR.Printf("%s is already submitted!\n", text)
				isexistTemp = true
				time.Sleep(2 * time.Second)
				break
			}
		}
		if isexistTemp {
			continue
		}
		tries = append(tries, text)
		//for used and found letters collections
		for _, try := range tries {
			for i, l := range try {
				found := false
				if try[i] == wordle[i] {
					matches[l] = true
					continue
				} else {
					for _, w := range wordle {
						if l == w {
							exists[l] = true
							found = true
							break
						}
					}
				}
				if found {
					continue
				}
				notExists[l] = true
			}
		}
	}
}

//replaces line breaks according to OS
func replacer(text string) string {

	old := ""
	if runtime.GOOS == "linux" {
		old = "\n"
	} else if runtime.GOOS == "windows" {
		old = "\r\n"
	}
	text = strings.Replace(text, old, "", -1)
	text = strings.ToLower(text)
	return text
}

//go:embed wordList.txt
var file embed.FS

//returns a 5-letter word
func returnRandomWord() (string, error) {

	read, err := file.ReadFile("wordList.txt")

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	word := []byte{}
	words := []string{}
	for _, b := range read {
		if unicode.IsLetter(rune(b)) {
			word = append(word, b)
			if len(word) == 5 {
				words = append(words, string(word))
				word = []byte{}
			}
		}
	}
	sort.Strings(words)
	Words = words
	rand.Seed(time.Now().UnixMilli())
	return strings.ToUpper(words[rand.Intn(len(words))]), nil
}

var Words []string

//checks if the word is valid
func isItaRealWord(str string) bool {
	fmt.Println(str, len(Words))
	for _, word := range Words {
		if word == str {
			return true
		}
	}
	return false
}

var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

//clear terminal
func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}
