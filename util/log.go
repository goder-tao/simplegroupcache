package util

import "log"

func Info(s string)  {
	log.Printf("[INFO] %s\n", s)
}

func Error(s string)  {
	log.Printf("[ERROR] %s\n", s)
}

func Fatal(s string)  {
	log.Fatalf("[FATAL] %s\n", s)
}