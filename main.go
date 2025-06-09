package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
)

//DA CREARE UN PUNTO DI INGRESSE PER I FUTURI AGGIORNAMENTI DELLA VM
//MAGARI QUALCHE METODO PER L'ESENSIONE E UN SISTEMA PER SOSTITUIRE COMPLETAMENTE QUESTO SERVIZIO CON UNO NUOVO IDENTIFICANDO QUESTO NUOVO COME LA NUOVA VM

func main() {
	var dirWatcher *fsnotify.Watcher
	var err error

	//I try 3 times to start watcher
	for i := 0; i < 3; i++ {
		dirWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		panic("Watcher not started: " + err.Error())
	}

	defer dirWatcher.Close()

	//I try 3 times to start compose dir to watcher
	dirToWatch := ""
	for i := 0; i < 3; i++ {
		baseDir := ""
		//get directory to watch
		switch runtime.GOOS {
		case "windows":
			if drive := os.Getenv("SystemDrive"); drive != "" {
				baseDir = drive + "/"
				break
			}
			baseDir = "C:/"
		case "linux", "android", "darwin":
			baseDir = "/"
		default:
			baseDir = "/"
		}
		dirToWatch, err = filepath.Abs(filepath.Join(baseDir, "GoChain", "Blockchain"))
		if err != nil {
			continue
		}

		// Create directory if it doesn't exist
		err = os.MkdirAll(dirToWatch, 0700) //only owner can modify
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		panic("Dir to watch cannot be initialized: " + err.Error())
	}

	//add directory to watcher, try 3 times
	for i := 0; i < 3; i++ {
		err = dirWatcher.Add(dirToWatch)
		if err != nil {
			continue
		}

		break
	}

	//go routine to actually watch dir
	go func() {
		for event := range dirWatcher.Events {
			//new file
			if event.Op&fsnotify.Write == fsnotify.Write {
				//SPOSTARE QUI LA LOGICA DI ASSEGNAZIONE DEL NOME DEI FILE DELLA BLOCKCHAIN
				//TESTARE SE è PRESENTE UN FILE .wasm DEL NUOVO FILE CREATO
			}
		}
	}()

}
