package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	md5simd "github.com/minio/md5-simd"
)

var server = md5simd.NewServer()

// Fonction pour afficher un chemin

func main() {

	// Liste des répertoires à parcourir

	roots, excludes := getOptions()
	// Parcourir l'arborescence en utilisant filepath.Walk
	for _, root := range strings.Split(roots, ",") {
		err := walk(root, excludes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erreur lors du parcours de l'arborescence : %v\n", err)
		}
	}
}

func getOptions() (dirPtr, exclude string) {
	// Définir un drapeau pour le chemin du répertoire
	flag.StringVar(&dirPtr, "d", "", "Chemin du répertoire à parcourir")
	flag.StringVar(&exclude, "e", "\\.git\\\\", "regexp des Liste de chemins à exclure")

	// Parser les arguments de ligne de commande
	flag.Parse()

	// Vérifier si le drapeau a été spécifié
	if dirPtr == "" {
		fmt.Fprintf(os.Stderr, "Veuillez spécifier un répertoire à l'aide du drapeau -d\n")
		return "", ""
	}

	// Vérifier si le répertoire existe
	if _, err := os.Stat(dirPtr); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Le répertoire spécifié n'existe pas.\n")
		return "", ""
	}
	return
}

func walk(dirPtr, exclude string) error {
	excludec, rerr := regexp.Compile(exclude)
	if rerr != nil {
		return rerr
	}
	return filepath.Walk(dirPtr, func(path string, info os.FileInfo, err error) error {
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
			matched := excludec.MatchString(path)
			if matched {
				return nil
			}
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
		fmt.Fprintf(os.Stderr, "Erreur lors de l'ouverture du fichier: %s %v", thePath, err)
		return -1, err
	}
	return fileInfo.Size(), err
}
