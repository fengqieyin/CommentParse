package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		printHelp()
		return
	}
	currentDate := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("%s.log", currentDate)
	// open log file
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("无法打开日志文件:", err)
		return
	}
	defer file.Close()

	// set log output
	log.SetOutput(file)

	dir := args[0]
	if err := countCommentLines(dir); err != nil {
		fmt.Println(err)
	}
}

func printHelp() {
	fmt.Println("usage: \n\tgo run . <directory>")
}

func countCommentLines(dir string) error {
	var wg sync.WaitGroup
	result := []*calculate{}
	var lock sync.Mutex
	// Recursively read files in a directory
	err := filepath.Walk(dir, func(fileName string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("An error occurred while accessing the file: %s\n", err.Error())
			return err
		}

		// Check if it is a file
		if !info.Mode().IsRegular() {
			return nil
		}
		concurrencyLimit := 5 // Maximum number of concurrent executions
		concurrencyLimitChan := make(chan struct{}, concurrencyLimit)
		// Check file extension
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == ".c" || ext == ".cpp" || ext == ".h" || ext == ".hpp" {
			concurrencyLimitChan <- struct{}{}
			wg.Add(1)
			go func(fileName string) {
				defer func() {
					<-concurrencyLimitChan
					wg.Done()
				}()

				// Read file contents
				fileContent, err := ioutil.ReadFile(fileName)
				if err != nil {
					log.Printf("打开文件时出错: %s\n", err.Error())
					return
				}

				// Count the number of comment lines
				cal := countCommentLinesInContent(fileContent)
				cal.name = fileName
				lock.Lock()
				result = append(result, cal)
				lock.Unlock()
			}(fileName)
		}

		return nil
	})

	if err != nil {
		log.Printf("An error occurred while accessing the directory: %s\n", err.Error())
		return err
	}

	// Wait for all coroutines to complete
	wg.Wait()
	sort.Slice(result, func(i int, j int) bool {
		return result[i].name < result[j].name
	})

	for _, item := range result {
		fmt.Printf("Name: %-40s total: %-4d inline : %-5d block : %-5d\n",
			item.name, item.total, len(item.inlineSet), len(item.blockSet))
	}

	return nil
}

func countSingleChar(content []byte, index int, cal *calculate, state *calState) {
	item := rune(content[index])
	pre := rune(0)
	if index > 0 {
		pre = rune(content[index-1])
	}

	// Among all states, whichever state is entered first shall prevail.

	// process the case of ' '
	if state.isInSingleQuote {
		if item == runeSinleQuote {
			if pre != runeAfterSlash {
				state.isInSingleQuote = false
			}
			// process '\\'
			if index > 1 && pre == runeAfterSlash {
				preTwo := rune(content[index-2])
				if preTwo == runeAfterSlash {
					state.isInSingleQuote = false
				}
			}
		}
		return
	}

	// R"()""
	if state.inRMode >= 0 {
		if item == runeQuote {
			i := 1
			isFind := false
			for {
				preMore := rune(content[index-i])
				if preMore == runeRightParentheses {
					isFind = true
					break
				}
				if index-i < state.inRMode {
					break
				}
				i++
			}
			if isFind {
				state.inRMode = -1
			}
		}
		return
	}

	// ""
	if state.isInquotes {
		if item == runeQuote {
			if pre != runeAfterSlash {
				state.isInquotes = false
			}
			// "\\\\"" 、"\\\""
			if pre == runeAfterSlash {
				i := 2
				preCnt := 1
				for index-i >= 0 {
					preMore := rune(content[index-i])
					if preMore != runeAfterSlash {
						break
					}
					i++
					preCnt++
				}
				if preCnt%2 == 0 {
					state.isInquotes = false
				}
			}
		}
		return
	}

	// /**/
	if state.isInblock {
		if item == runeSlash && pre == runeStart {
			state.isInblock = false
			cal.blockSet[cal.total] = struct{}{}
		}
		return
	}

	// Determine which state to enter based on the current character
	if item == runeSinleQuote {
		state.isInSingleQuote = true
		return
	}

	if item == runeQuote {
		if pre == runeR {
			state.inRMode = index
		} else {
			state.isInquotes = true
		}
		return
	}

	if item == runeStart && pre == runeSlash {
		state.isInblock = true
		return
	}

	// //
	if item == runeSlash && pre == runeSlash {
		cal.inlineSet[cal.total] = struct{}{}
		state.isGotoEnd = true
		return
	}
}

func countCommentLinesInContent(content []byte) *calculate {
	// Initialize
	cal := new(calculate)
	state := new(calState)
	cal.inlineSet = make(map[int]struct{})
	cal.blockSet = make(map[int]struct{})
	state.isInblock = false
	state.isInquotes = false
	state.isInSingleQuote = false
	state.isGotoEnd = false
	state.inRMode = -1

	for i := range content {
		item := rune(content[i])
		// Skip if state is in go-to-end mode and the character is not a runeChange
		if state.isGotoEnd && item != runeChange {
			continue
		}

		if item == runeChange {
			if state.isInblock {
				cal.blockSet[cal.total] = struct{}{}
			}
			pre := rune(0)
			if i > 0 {
				pre = rune(content[i-1])
			}

			// Process inline comment witch adding a line with the right slash
			if state.isGotoEnd && pre == runeAfterSlash {
				cal.inlineSet[cal.total+1] = struct{}{}
			} else {
				state.isGotoEnd = false
			}
			// Increment the total count
			cal.total++
			continue
		}

		countSingleChar(content, i, cal, state)
	}

	return cal
}
