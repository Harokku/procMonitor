package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {

	var (
		monitoredList []string // Process exe to monitor
		maxMemory     string   // Max allowed memory usage in KB
		checkinterval int      // Process check interval in minute3s
		err           error
	)

	// -------------------------
	// .env loading
	// -------------------------

	// Load YAML with godotenv pkg
	err = godotenv.Load("env.yaml")

	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Parse yaml and build monitor data
	for _, s := range strings.Split(os.Getenv("MONITOREDLIST"), "|") {
		monitoredList = append(monitoredList, s)
	}
	maxMemory = os.Getenv("MAXMEMORY")
	checkinterval, err = strconv.Atoi(os.Getenv("CHECKINTERVAL"))
	if err != nil {
		log.Fatalf("Error parsing check interval in yaml file: %s", err)
	}

	// Log info about monitored process for user
	log.Println("Monitored process list:")
	fmt.Println("N\tProcess name")
	for i, process := range monitoredList {
		fmt.Printf("%d\t%s\n", i, process)
	}

	// Cycle till interrupted, monitoring processes
	for {
		// For each monitored process check memory usage and if too high kill it
		for _, process := range monitoredList {
			// Invoke win tasklist to check for specified process by process name
			cmd := exec.Command("tasklist",
				"/FI", fmt.Sprintf("IMAGENAME eq %s", process),
				"/FI", fmt.Sprintf("MEMUSAGE gt %s", maxMemory),
				"/FO", "CSV",
				"/NH")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("Error invoking tasklist", string(out), err)
			}

			// Check if process is found and if is overusing memory
			// Check if output start with ERROR: or INFO:
			errorCheck := strings.Split(string(out), ":")
			if errorCheck[0] != "ERROR" && errorCheck[0] != "INFO" {
				// Process is consuming too much memory, kill it
				// Invokes win taskkill to forcefully end the process and his children
				log.Printf("MEMORY LEAK:\t%s", out)
				log.Printf("Process %s\tconsuming too much memory... Killing it", process)
				cmd := exec.Command("taskkill",
					"/IM", process,
					"/F",
					"/T")
				out, err := cmd.CombinedOutput()
				if err != nil {
					log.Println("Error killing process", err)
				}
				// Check if output is success
				errorCheck = strings.Split(string(out), ":")
				if errorCheck[0] == "SUCCESS" {
					log.Println("Process correctly killed")
				} else {
					log.Println("Error killing process: ", err)
				}
			}
		}
		time.Sleep(time.Duration(checkinterval) * time.Minute)
	}
}
