package main

import "strings"

func cleanChirp(chirpOrig string) (bool, string) {
	words := strings.Split(chirpOrig, " ")
	containsProfanity := false
	for i := 0; i < len(words); i++ {
		word := strings.ToLower(words[i])
		switch word {
		case "kerfuffle", "sharbert", "fornax":
			containsProfanity = true
			words[i] = "****"
		}
	}
	cleanedChirp := strings.Join(words, " ")
	return containsProfanity, cleanedChirp
}
