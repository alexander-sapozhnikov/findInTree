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
	// слово, которое ищем
	word string
	// формально очередь конкурентная для ответов
	results chan string

	// синхронизация между потоками
	wgExecutor = &sync.WaitGroup{}
	wqReader   = &sync.WaitGroup{}

	// синхронизация внутри потока
	mutex = &sync.Mutex{}

	// считаем через очередь сколько горутин должно быть запущено
	// так имитируем пул горутин. В го пула нет.
	activeGoroutines chan struct{}
	// кол-во горутин, которые передали при запуске
	countGoroutines int
)

func main() {
	// берем все передаваемые аргументы
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

	// инициализируем очереди
	results = make(chan string, 10000)
	activeGoroutines = make(chan struct{}, countGoroutines*2)

	startTime := time.Now()
	// помечаем, что сейчас один поток блокирует основой
	wgExecutor.Add(1)
	// отмечаем, что запущена одна горутина
	activeGoroutines <- struct{}{}
	// запускаем поиск
	searchFiles(path, true)

	// помечаем, что поток чтения работает
	wqReader.Add(1)
	var ans []string
	go func() {
		// читаем с очереди все найденные файлы
		// когда закрываем results, то происходит выход из цикла
		for file := range results {
			ans = append(ans, file)
		}
		// снимаем блокировку с основного потока
		wqReader.Done()
	}()

	// ждем пока все файлы обойдем
	wgExecutor.Wait()
	// закрываем очередь результатов
	close(results)
	// ждем завершение вывода найденных файлов
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
		filename := file.Name()
		if strings.Contains(filename, word) {
			// названия удовлетворяет заданному слову, поэтому кладем его в ответ
			results <- filename
		}

		if file.IsDir() {
			mutex.Lock()
			pathNew := filepath.Join(path, file.Name())
			if len(activeGoroutines) == countGoroutines {
				// не можем создать новый потом т.к. все заняты
				// работаем в текущем
				mutex.Unlock()
				searchFiles(pathNew, false)
			} else {
				// создаем новый поток
				activeGoroutines <- struct{}{}
				wgExecutor.Add(1)
				mutex.Unlock()
				go searchFiles(pathNew, true)
			}
		}
	}

	if isNew {
		gr := bytes.Fields(debug.Stack())[1]
		fmt.Printf("End goorutine  #%v and path = %v\n", string(gr), path)
		// уменьшаем кол-во запущенных потоков
		<-activeGoroutines
		// снимаем блокировку с основного потока
		wgExecutor.Done()
	}

}
