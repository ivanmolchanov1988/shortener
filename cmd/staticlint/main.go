package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"

	"honnef.co/go/tools/staticcheck"

	"github.com/ivanmolchanov1988/shortener/cmd/staticlint/checkers"
)

// Config — имя файла конфигурации.
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string `json:"staticcheck"`
}

func main() {
	fmt.Println("Запуск кастомного статического анализатора...")

	// Загружаем конфигурационный файл
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}

	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	// Подключаем стандартные анализаторы
	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		fieldalignment.Analyzer,
		checkers.NoOsExitAnalyzer,
	}

	// Добавляем staticcheck-анализаторы из конфигурации
	checks := make(map[string]bool)
	for _, v := range cfg.Staticcheck {
		checks[v] = true
	}

	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// Запускаем multichecker
	multichecker.Main(mychecks...)
}
