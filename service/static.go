package service

import (
	"log"
	"net/http"
)

// RunStaticTorServer runs a static file server over Tor
func RunStaticTorServer(staticDir, vanityName string) {
	log.Printf("Starting static file server on port 8080 serving directory: %s", staticDir)
	runTorHiddenService(vanityName, func() {
		handler := http.FileServer(http.Dir(staticDir))
		http.Handle("/", handler)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("Static server error: %v", err)
		}
	})
}
