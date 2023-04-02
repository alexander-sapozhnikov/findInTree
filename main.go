package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	word             string
	results          chan string

	wgExecutor       = &sync.WaitGroup{}
	wqReader         = &sync.WaitGroup{}


	activeGoroutines chan struct{}
	countGoroutines  int
)

func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		fmt.Println("Запусти программу так: go run main.go <path> <word> [countGoroutines]")
		return
	}

	path := args[0]
	word = args[1]
	countGoroutines = 1
	if len(args) == 3 {
		var err error
		countGoroutines, err = strconv.Atoi(args[2])
		if countGoroutines <= 0 || err != nil {
			fmt.Println("Некорректное число горутин")
			return
		}
	}

	results = make(chan string, countGoroutines*2)
	activeGoroutines = make(chan struct{}, countGoroutines*2)

	startTime := time.Now()
	wgExecutor.Add(1)
	activeGoroutines <- struct{}{}
	searchFiles(path, true)

	wqReader.Add(1)
	var ans []string
	go func() {
		for file := range results {
			ans = append(ans, file)
		}
		wqReader.Done()
	}()

	wgExecutor.Wait()
	close(results)
	wqReader.Wait()

	fmt.Printf("Итоговое время: %v\n", time.Since(startTime))
	fmt.Printf("Нашёл файлы: %v\n", ans)
}


// Функция для поиска файлов с заданным словом
// isNew - функция запустилась в новом потоке
func searchFiles(path string, isNew bool) {
	if isNew {
		gr := bytes.Fields(debug.Stack())[1]
		fmt.Printf("Start goorutine  #%v and path = %v\n", string(gr), path)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			if strings.Contains(filename, word) {
				results <- filename
			}
		} else {
			pathNew := filepath.Join(path, file.Name())
			if len(activeGoroutines) == countGoroutines {
				// не можем создать новый потом т.к. все заняты
				// работаем в текущем
				searchFiles(pathNew, false)
			} else {
				// создаем новый поток
				activeGoroutines <- struct{}{}
				wgExecutor.Add(1)
				go searchFiles(pathNew, true)
			}
		}
	}

	if isNew {
		gr := bytes.Fields(debug.Stack())[1]
		fmt.Printf("End goorutine  #%v and path = %v\n", string(gr), path)
		<-activeGoroutines
		wgExecutor.Done()
	}

}
