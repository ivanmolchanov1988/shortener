// Package server управляет инициализацией конфигурации и запуском HTTP-сервера.
// Утилиты для server.
package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Generic-функция для проверки пустых значений и подстановки "N/A"
func setDefaultNA[T comparable](value *T, defaultValue T) {
	var zero T // переменная типа T с nil значением
	if *value == zero {
		*value = defaultValue
	}
}

// Для автотестов Яндекса
func copyMigrations(srcDir, dstDir string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source migrations directory: %v", err)
	}
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		err = os.MkdirAll(dstDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create destination migrations directory: %v", err)
		}
	}
	for _, file := range files {
		srcFilePath := filepath.Join(srcDir, file.Name())
		dstFilePath := filepath.Join(dstDir, file.Name())

		srcFile, err := os.Open(srcFilePath)
		if err != nil {
			return fmt.Errorf("failed to open source file: %v", err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstFilePath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %v", err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}
	return nil
}

// CreateDirectories создаёт каталоги, хз для чего, не помню, 5 часов утра.
func CreateDirectories(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist, creating: %v\n", dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		fmt.Printf("Directory created: %v\n", dir)
	} else if err != nil {
		return fmt.Errorf("error checking directory: %w", err)
	} else {
		fmt.Printf("Directory already exists: %v\n", dir)
	}
	return nil
}

// CreateFileIfNotExist - Проверяем наличие файла и создаем его, если он отсутствует.
func CreateFileIfNotExist(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist, creating: %v\n", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		file.Close()
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	} else {
		fmt.Printf("File already exists: %v\n", filePath)
	}
	return nil
}

/// help funcs - вынести из server

func getProjectRoot() string {
	// Используем текущий рабочий каталог как корневой каталог
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return ""
	}
	return dir
}

//	func getDefaultFilePath() string {
//		projectRoot := getProjectRoot()
//		newPath := filepath.Join(projectRoot, "urls.json")
//		return newPath
//	}
func getShortRoot() string {
	var fullRoot = getProjectRoot()
	index := strings.Index(fullRoot, "/cmd/shortener")
	if index == -1 {
		fmt.Println("Failed to find ShortRoot")
		return fullRoot
	}
	return fullRoot[:index]
}

// firstNonEmpty - выбирает первое непустое значение.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// firstNonEmptyBool - выбирает первое значение bool, если оно задано.
func firstNonEmptyBool(values ...bool) bool {
	for _, v := range values {
		if v {
			return v
		}
	}
	return false
}

// parseBool - преобразует строку в bool.
func parseBool(value string) bool {
	return value == "true"
}
