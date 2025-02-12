// Package server управляет инициализацией конфигурации и запуском HTTP-сервера.
// Для инициализации хранилища в sever.
package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ivanmolchanov1988/shortener/internal/filestore"
)

func initializeDatabase(dbDSN string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbDSN)
	if db == nil || err != nil {
		return nil, errors.New("db connection is nil or return erro")
	}

	// ----- Для автотестов
	// Определение исходного и целевого путей для миграций
	srcMigrationsPath := filepath.Join(getProjectRoot(), "internal/migrations")

	// Определяем директорию запуска тестов
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}
	dstMigrationsPath := filepath.Join(currentDir, "internal/migrations")

	// Копируем миграции в нужную директорию, если они там отсутствуют
	if _, err := os.Stat(dstMigrationsPath); os.IsNotExist(err) {
		if err := copyMigrations(srcMigrationsPath, dstMigrationsPath); err != nil {
			return nil, fmt.Errorf("failed to copy migrations: %v", err)
		}
	}
	fmt.Printf("root project migrations folder: %v\n", srcMigrationsPath)
	fmt.Printf("dst migrations folder: %v\n", dstMigrationsPath)
	// -----

	// миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate: %v", err)
	}
	rootPath := getShortRoot()
	files, err := os.ReadDir(rootPath + "/internal/migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %v", err)
	}
	for _, file := range files {
		log.Printf("found migration file: %s", file.Name())
	}
	m, err := migrate.NewWithDatabaseInstance(
		//"file://"+rootPath+"/internal/migrations",
		"file://"+dstMigrationsPath, // - ДЛЯ ЯНДЕКСА
		baseDSN.dbname,
		driver)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrate: %v", err)
	}
	//запуск миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrate: %v", err)
	}
	log.Println("Migration is complete!")

	return db, nil
}

func initializeFileStorage(filePath string) (*filestore.FileStorage, error) {
	fmt.Printf("Starting to initialize file storage at %s\n", filePath)

	dirPath := filepath.Dir(filePath)
	if err := CreateDirectories(dirPath); err != nil {
		fmt.Printf("Failed to create directories for path %s: %v\n", dirPath, err)
		return nil, err
	}

	if err := CreateFileIfNotExist(filePath); err != nil {
		return nil, err
	}

	store, err := filestore.NewFileStorage(filePath)
	if err != nil {
		return nil, err
	}

	return store, nil
}
