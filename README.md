# loc-counter - Lines of Code Counter

Утилита считает строки **реального кода** в исходных файлах, пропуская:
- Пустые строки
- Строки, целиком занятые комментариями (`//`, `#`, `/* ... */`)
- Строки внутри многострочных блочных комментариев

Строки, содержащие **код и комментарий одновременно** — учитываются.

## Готовые сборки

Исполняемые файлы для самых популярных платформ доступны на странице [Releases](https://github.com/alex6712/loc-counter/releases).
Вы можете скачать подходящую версию и сразу начать использовать:

- **Linux** ([amd64](https://github.com/alex6712/loc-counter/releases/download/v0.1.3/loc_counter_v0.1.3_linux_amd64.tar.gz), [arm64](https://github.com/alex6712/loc-counter/releases/download/v0.1.3/loc_counter_v0.1.3_linux_arm64.tar.gz))
- **Windows** ([amd64](https://github.com/alex6712/loc-counter/releases/download/v0.1.3/loc_counter_v0.1.3_windows_amd64.zip))

Просто распакуйте архив и запустите исполняемый файл из терминала / командной строки.

## Поддерживаемые расширения

| Язык            | Расширения                        |
|-----------------|-----------------------------------|
| C               | `.c`, `.h`                        |
| C++             | `.cpp`, `.cc`, `.cxx`, `.hpp`     |
| Java            | `.java`                           |
| JavaScript      | `.js`, `.jsx`                     |
| TypeScript      | `.ts`, `.tsx`                     |
| Go              | `.go`                             |
| Rust            | `.rs`                             |
| C#              | `.cs`                             |
| Python          | `.py`                             |

## Сборка из исходников

Если вы хотите собрать утилиту самостоятельно:

**Linux / macOS:**
```bash
go build -o loc_counter .
```

**Windows** (расширение `.exe` обязательно, иначе Go создаст файл без него):
```powershell
go build -o loc_counter.exe .
```

## Использование

> На Windows замените префикс `./loc_counter` на `.\loc_counter` или `loc_counter.exe`.

```bash
# Подсчитать все поддерживаемые файлы в директории ./src
./loc_counter ./src

# Без аргументов — попросит ввести директорию (по умолчанию ".")
./loc_counter

# Учитывать только .go и .rs файлы
./loc_counter --ext .go --ext .rs ./src

# Исключить .py файлы
./loc_counter --ext-exclude .py ./src

# --ext-exclude имеет приоритет над --ext
# .py будет проигнорирован, даже если указан в --ext
./loc_counter --ext .go,.py --ext-exclude .py ./src

# Исключить директории (через запятую или отдельными флагами)
./loc_counter --exclude .venv,node_modules,.git ./src
./loc_counter --exclude .venv --exclude node_modules ./src

# Исключить конкретную вложенную директорию
./loc_counter --exclude internal/generated ./src
```

## Добавление нового языка

В файле `main.go` найдите переменную `knownLanguages` и добавьте запись:

```go
".rb": {
    SingleLine: []string{"#"},
    MultiStart: "=begin",
    MultiEnd:   "=end",
},
```

Для языков с C-стилем комментариев используйте готовую функцию `cStyleConfig()`.
