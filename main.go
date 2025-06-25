package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	baseDirName           = "GoChain"
	blockchainDir         = "Blockchain"
	blockchainLogDir      = "BlockchainLog"
	logFileName           = "chainLog.log"
	binExtension          = ".bin"
	expectedFileLenNoNext = 128 //length of file without the next hash file in file name
	fileHashLength        = 64
	pollingInterval       = time.Second // Intervallo di polling: controlla ogni secondo
)

//DA CREARE UN PUNTO DI INGRESSE PER I FUTURI AGGIORNAMENTI DELLA VM
//MAGARI QUALCHE METODO PER L'ESENSIONE E UN SISTEMA PER SOSTITUIRE COMPLETAMENTE QUESTO SERVIZIO CON UNO NUOVO IDENTIFICANDO QUESTO NUOVO COME LA NUOVA VM

func main() {
	var logFile *os.File
	done := make(chan struct{})

	//get the base directory
	baseDir := getBaseDir()

	//setup log file and get it. If function returns the log file exists
	logFile = setupLogging(baseDir)
	//close log file at the end of function
	defer logFile.Close()

	//CON GO STANDARD POSSO USARE IL WATCHER
	//setup watcher and get it. If function returns the watcher exists
	dirWatcher := setupWatcher()
	// close watcher dir at the end of function
	defer dirWatcher.Close()

	//setup directory to watch
	dirToWatch := ""
	setupDirToWatch(baseDir, &dirToWatch, dirWatcher)
	// setupDirToWatch(baseDir, &dirToWatch) //NON USIAMO IL WATCHER MA IL POLLING PER ORA
	log.Printf("dir to watch: " + dirToWatch)

	var wg sync.WaitGroup
	wg.Add(1)

	//go routine to actually watch dir events
	go func() {
		log.Printf("Cycle to monitor directory")
		defer wg.Done()
		for {
			// function to polling process directory's files ù
			// MA CON GO STANDARD POSSO USARE IL WATCHER
			// processFilesInDirectory(dirToWatch)
			// time.Sleep(pollingInterval)

			//CON GO STANDARD POSSO USARE IL WATCHER
			select {
			case err, ok := <-dirWatcher.Errors:
				if !ok {
					log.Printf("watcher errors not ok")
					return // exit from for loop
				}
				log.Printf("Watcher error: %v", err)

			case event, ok := <-dirWatcher.Events:
				if !ok {
					log.Printf("watcher event not ok")
					return // exit from for loop
				}
				processWatcherEvent(event, dirToWatch)
			}
		}
	}()

	//block program until done will be closed
	<-done

	//PER ORA NON USIAMO IL WATCHER MA IL POLLING
	// close dirWatcher to let go routine exit for cycle of watcher events
	// dirWatcher.Close()

	//wait for go routine end
	wg.Wait()
	log.Printf("Program ended")
}

func getBaseDir() string {
	switch runtime.GOOS {
	case "windows":
		if drive := os.Getenv("SystemDrive"); drive != "" {
			return drive + "/"
		}
		return "C:/"
	case "linux", "android", "darwin":
		return "/"
	default:
		return "/"
	}
}

// Create log file setting up flags and returns its pointer
func setupLogging(baseDir string) *os.File {
	var logDir string
	var logFilePath string
	var logFile *os.File
	var err error

	for i := 0; i < 3; i++ {
		logDir, err = filepath.Abs(filepath.Join(baseDir, "GoChain", "BlockchainLog"))
		if err != nil {
			log.Printf("Cannot create path for log")
			continue
		}
		logFilePath = filepath.Join(logDir, "chainLog.log")

		// Create directory if it doesn't exist
		err = os.MkdirAll(logDir, 0644) //only owner can modify
		if err != nil {
			log.Printf("Cannot create dir for log")
			continue
		}
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Cannot open file for log")
			continue
		}

		//set path file log as file to write log
		log.SetOutput(logFile)
		//set default log information
		log.SetFlags(log.Ldate | log.Ltime)
		break
	}
	if err != nil || logFilePath == "" {
		log.Panic("Cannot create log file: " + logFilePath)
	}

	return logFile
}

// CON GO STANDARD POSSO USARE IL WATCHER
// Create watcher and returns its pointer
func setupWatcher() *fsnotify.Watcher {
	var dirWatcher *fsnotify.Watcher
	var err error

	for i := 0; i < 3; i++ {
		dirWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			log.Printf("Cannot create watcher")
			continue
		}
		break
	}
	if err != nil {
		log.Panic("Watcher not started: " + err.Error())
	}
	return dirWatcher
}

// setup directory to watch adding it to watcher
func setupDirToWatch(baseDir string, dirToWatch *string, watcher *fsnotify.Watcher) { // CON GO STANDARD POSSO USARE IL WATCHER
	// func setupDirToWatch(baseDir string, dirToWatch *string) { //non usiamo il watcher ma il polling per ora
	var err error

	for i := 0; i < 3; i++ {
		*dirToWatch, err = filepath.Abs(filepath.Join(baseDir, "GoChain", "Blockchain"))
		if err != nil {
			log.Printf("Cannot get absolute path of dir to watch")
			continue
		}
		if _, err := os.Stat(*dirToWatch); os.IsNotExist(err) {
			log.Panic("Directory to watch %s doesnt' exitsts", *dirToWatch)
		}
		break
	}
	if err != nil {
		log.Panic("Dir to watch cannot be initialized: " + err.Error())
	}

	//CON GO STANDARD POSSO USARE IL WATCHER
	// add directory to watcher, try 3 times
	for i := 0; i < 3; i++ {
		err = watcher.Add(*dirToWatch)
		if err != nil {
			log.Printf("Cannot watch dir: %s", *dirToWatch)
			continue
		}

		break
	}
}

// //CON GO STANDARD POSSO USARE IL WATCHER
// manage watcher's event
func processWatcherEvent(event fsnotify.Event, dirToWatch string) {
	log.Printf("Event from watcher: %s, Op: %s", event.Name, event.Op.String())

	//new file is created
	if event.Op&fsnotify.Create == fsnotify.Create {
		log.Printf("Find new file: %s", event.Name)
	}

	//a file is written
	if event.Op&fsnotify.Write == fsnotify.Write {
		log.Printf("Wrote new file: %s", event.Name)

		manageFileRenaming(event.Name, dirToWatch)

		//TESTARE SE è PRESENTE UN FILE .wasm DEL NUOVO FILE CREATO
	}
}

// This function accept the new file name as parameter and cycling files in dirToWatch looks for the previous file thanks to hash of prev file in the first 64 bytes of new file name
func manageFileRenaming(filePath string, dirToWatch string) {
	fileFullName := filepath.Base(filePath)
	extFile := filepath.Ext(fileFullName)

	// Considera solo i file con estensione .bin
	if extFile != binExtension {
		return
	}

	//first 64 chars of fileName are the hash of previous transaction that is the file I am looking for
	fileName := strings.TrimSuffix(filepath.Base(fileFullName), extFile)
	if len(fileName) < fileHashLength {
		log.Printf("File name not in correct format")
		return
	}

	log.Printf("New file found: %s", fileFullName)

	var fileNameToUpdate string
	//the file's name of event cannot be prefix by O- cause O- is the prefix ONLY for Genesys block that is the start of all vm
	prevHashFile := fileName[0:fileHashLength]
	hashFile := fileName[len(fileName)-fileHashLength:]

	//Look for the file whose name is composed only by prevHash + fileHash or by O-fileHash that indicates the last transaction
	files, err := os.ReadDir(dirToWatch)
	if err != nil {
		log.Printf("Error on read transacionts' directory: %v", err)
	} else {
		for _, f := range files {

			//ignore dir and file not .bin
			if f.IsDir() || filepath.Ext(f.Name()) != binExtension {
				continue
			}

			//get file name without extension
			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

			//the length of previous file name must be of 128 bytes because it doesn't have the next file hash yet
			if len(name) > fileHashLength {
				if (len(name) == expectedFileLenNoNext && name[fileHashLength:len(name)] == prevHashFile) || (len(name) == 66 && name == "O-"+prevHashFile) {
					// 192 hex chars is transaction block with another block next
					// 128 hex chars is transaction block without another block next
					// 66 hex chars pre-pended by "O-" is genesys block without another block next
					fileNameToUpdate = name
					break
				}
			}
		}
	}

	if fileNameToUpdate == "" {
		log.Printf("File to update not found, hash: %s", prevHashFile)
	}
	err = os.Rename(filepath.Join(dirToWatch, fileNameToUpdate+".bin"), filepath.Join(dirToWatch, fileNameToUpdate+hashFile+".bin"))
	if err != nil {
		log.Printf("Error on file renaming: %s", fileNameToUpdate+".bin")
	}
}

//FUNZIONE PER POLLING, MI FA UN PO' SCHIFO MA PER ORA NON USIAMO IL WATCHER PERCHè TINYGO NON LO SUPPORTA
// processFilesInDirectory scansiona la directory e chiama processFileLogic per i file pertinenti.
// func processFilesInDirectory(dirToWatch string) {
// 	// log.Printf("Scanning directory: %s\r\n", dirToWatch) // Commentato per ridurre il log verboso

// 	files, err := os.ReadDir(dirToWatch)
// 	if err != nil {
// 		log.Panicf("Error reading directory %s: %v\r\n", dirToWatch, err)
// 		return
// 	}

// 	for _, f := range files {
// 		if f.IsDir() {
// 			continue
// 		}

// 		manageFileRenaming(f.Name(), dirToWatch);
// 	}
// }
