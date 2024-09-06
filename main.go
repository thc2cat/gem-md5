package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	md5simd "github.com/minio/md5-simd"
)

var server = md5simd.NewServer()

// Fonction pour afficher un chemin

func main() {

	// Définir un drapeau pour le chemin du répertoire
	dirPtr := flag.String("d", "", "Chemin du répertoire à parcourir")

	// Parser les arguments de ligne de commande
	flag.Parse()

	// Vérifier si le drapeau a été spécifié
	if *dirPtr == "" {
		fmt.Fprintf(os.Stderr, "Veuillez spécifier un répertoire à l'aide du drapeau -d\n")
		return
	}

	// Vérifier si le répertoire existe
	if _, err := os.Stat(*dirPtr); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Le répertoire spécifié n'existe pas.\n")
		return
	}

	// Parcourir l'arborescence en utilisant filepath.Walk
	err := filepath.Walk(*dirPtr, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Le fichier %s n'existe plus.\n", path)
				return nil // On ignore cette erreur et on continue
			} else if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Accès refusé au répertoire : %s\n", path)
				return nil // On ignore cette erreur et on continue
			} else {
				return err // On propage les autres types d'erreurs
			}
		}

		if !info.IsDir() {
			size, serr := getSize(path)
			if serr != nil || size == 0 {
				return nil
			}
			S, E := getHash(path)
			if E == nil {
				fmt.Printf("%s:%d:%s\n", S, size, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors du parcours de l'arborescence : %v\n", err)
	}
}

func getHash(thePath string) (string, error) {

	h := server.NewHash()
	defer h.Reset()

	fd, err := os.Open(thePath)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	io.Copy(h, fd)

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getSize(thePath string) (int64, error) {
	fileInfo, err := os.Stat(thePath)
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier:", err)
		fmt.Fprintf(os.Stderr, "Erreur lors de l'ouverture du fichier: %s %v", thePath, err)
		return -1, err
	}
	return fileInfo.Size(), err
}
