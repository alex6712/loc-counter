package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LangConfig описывает синтаксис комментариев для языка.
type LangConfig struct {
	SingleLine []string // префиксы, обозначающие начало однострочного или inline-комментария
	MultiStart string   // начало блочного комментария
	MultiEnd   string   // конец блочного комментария
}

// knownLanguages сопоставляет расширение файла и конфигурацию языка.
// Чтобы добавить новый язык, просто добавьте сюда новую запись.
var knownLanguages = map[string]LangConfig{
	// C-подобные языки
	".c":   cStyleConfig(),
	".h":   cStyleConfig(),
	".cpp": cStyleConfig(),
	".cc":  cStyleConfig(),
	".cxx": cStyleConfig(),
	".hpp": cStyleConfig(),
	// Java
	".java": cStyleConfig(),
	// JavaScript / TypeScript
	".js":  cStyleConfig(),
	".ts":  cStyleConfig(),
	".jsx": cStyleConfig(),
	".tsx": cStyleConfig(),
	// Go
	".go": cStyleConfig(),
	// Rust
	".rs": cStyleConfig(),
	// C#
	".cs": cStyleConfig(),
	// Python — нет отдельного токена блочного комментария,
	// используется # и тройные кавычки (обрабатываются как строки)
	".py": {
		SingleLine: []string{"#"},
		MultiStart: `"""`,
		MultiEnd:   `"""`,
	},
}

func cStyleConfig() LangConfig {
	return LangConfig{
		SingleLine: []string{"//"},
		MultiStart: "/*",
		MultiEnd:   "*/",
	}
}

// countLines подсчитывает логические строки кода в файле:
//   - Пустые строки пропускаются.
//   - Строки, полностью находящиеся внутри блочного комментария, пропускаются.
//   - Строки, содержащие только однострочный комментарий (после Trim), пропускаются.
//   - Строки, содержащие код И комментарий (inline), учитываются.
func countLines(path string, cfg LangConfig) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	inBlock := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// ---- Обработка блочных комментариев ----
		if cfg.MultiStart != "" {
			if inBlock {
				// Всё ещё внутри блочного комментария — проверяем конец
				if idx := strings.Index(trimmed, cfg.MultiEnd); idx >= 0 {
					inBlock = false
					// есть ли что-то после закрывающего токена на этой же строке?
					rest := strings.TrimSpace(trimmed[idx+len(cfg.MultiEnd):])
					if rest != "" && !isEntirelyComment(rest, cfg) {
						count++
					}
				}
				// В любом случае, эта строка не добавляет кода
				continue
			}

			// Не в блоке — проверяем, начинается ли блок здесь
			if startIdx := strings.Index(trimmed, cfg.MultiStart); startIdx >= 0 {
				// Есть ли код перед началом блока?
				before := strings.TrimSpace(trimmed[:startIdx])
				hasCodeBefore := before != "" && !isEntirelyComment(before, cfg)

				// Закрывается ли блок на этой же строке?
				searchFrom := startIdx + len(cfg.MultiStart)
				if endIdx := strings.Index(trimmed[searchFrom:], cfg.MultiEnd); endIdx >= 0 {
					// Однострочный блочный комментарий: /*...*/
					afterEnd := strings.TrimSpace(trimmed[searchFrom+endIdx+len(cfg.MultiEnd):])
					hasCodeAfter := afterEnd != "" && !isEntirelyComment(afterEnd, cfg)
					if hasCodeBefore || hasCodeAfter {
						count++
					}
					// inBlock остаётся false
					continue
				}

				// Блок начинается и НЕ заканчивается на этой строке
				inBlock = true
				if hasCodeBefore {
					count++
				}
				continue
			}
		}

		// ---- Не в блоке и блок не начинается на этой строке ----
		// Проверяем, покрывает ли однострочный комментарий всю строку
		if isEntirelyComment(trimmed, cfg) {
			continue
		}

		count++
	}

	return count, scanner.Err()
}

// isEntirelyComment возвращает true, если строка (после Trim)
// начинается с одного из токенов однострочного комментария.
func isEntirelyComment(s string, cfg LangConfig) bool {
	for _, prefix := range cfg.SingleLine {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// --- Вспомогательный тип флага StringSlice (позволяет использовать
// --ext .go --ext .py  ИЛИ  --ext .go,.py) ---

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error {
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*s = append(*s, normalizeExt(part))
		}
	}
	return nil
}

func normalizeExt(ext string) string {
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}

func main() {
	var extFlag stringSlice
	var excludeFlag stringSlice

	flag.Var(&extFlag, "ext", "Расширения для включения (например, --ext .go --ext .py). По умолчанию: все поддерживаемые.")
	flag.Var(&excludeFlag, "exclude", "Расширения для исключения (например, --exclude .py). Имеет приоритет над --ext.")
	flag.Parse()

	// Определяем директорию
	dir := ""
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	} else {
		fmt.Print("Введите путь к директории [.]: ")
		fmt.Scanln(&dir)
		if strings.TrimSpace(dir) == "" {
			dir = "."
		}
	}

	// Формируем набор исключений
	excludeSet := make(map[string]bool)
	for _, e := range excludeFlag {
		excludeSet[e] = true
	}

	// Формируем набор включений (nil означает «все поддерживаемые»)
	var includeSet map[string]bool
	if len(extFlag) > 0 {
		includeSet = make(map[string]bool)
		for _, e := range extFlag {
			includeSet[e] = true
		}
	}

	// Обход директории
	type fileResult struct {
		path  string
		lines int
	}

	var results []fileResult
	totalLines := 0
	totalFiles := 0

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "предупреждение: невозможно получить доступ к %s: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		cfg, supported := knownLanguages[ext]
		if !supported {
			return nil
		}

		// Применяем фильтры
		if excludeSet[ext] {
			return nil
		}
		if includeSet != nil && !includeSet[ext] {
			return nil
		}

		lines, err := countLines(path, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "предупреждение: невозможно прочитать %s: %v\n", path, err)
			return nil
		}

		results = append(results, fileResult{path, lines})
		totalLines += lines
		totalFiles++
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка обхода директории: %v\n", err)
		os.Exit(1)
	}

	if totalFiles == 0 {
		fmt.Println("Поддерживаемые исходные файлы не найдены.")
		return
	}

	// Вывод результатов по каждому файлу
	maxPathLen := 0
	for _, r := range results {
		if len(r.path) > maxPathLen {
			maxPathLen = len(r.path)
		}
	}

	fmt.Println()
	fmt.Printf("%-*s  %s\n", maxPathLen, "Файл", "Строки")
	fmt.Println(strings.Repeat("-", maxPathLen+10))
	for _, r := range results {
		fmt.Printf("%-*s  %d\n", maxPathLen, r.path, r.lines)
	}
	fmt.Println(strings.Repeat("-", maxPathLen+10))
	fmt.Printf("%-*s  %d\n", maxPathLen, fmt.Sprintf("Итого (%d файлов)", totalFiles), totalLines)
	fmt.Println()
}
