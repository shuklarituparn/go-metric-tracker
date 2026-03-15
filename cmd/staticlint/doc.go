// Package main содержит staticlint.
//
// # Запуск
//
// В режиме "как бинарь":
//
//	go build -o ./bin/staticlint ./cmd/staticlint
//	go vet -vettool=./bin/staticlint ./...
//
// # Что проверяет
//
// staticlint объединяет:
//
//   - стандартные анализаторы из golang.org/x/tools/go/analysis/passes
//     (ошибки в defer, printf, shadowing, unreachable и т.п.).
//
//   - все анализаторы класса SA из honnef.co/go/tools/staticcheck
//     (поиск потенциальных багов и опасных конструкций).
//
//   - анализаторы других классов staticcheck (S/ST/QF/U), чтобы покрыть
//     упрощения, стиль, quickfix и неиспользуемый код.
//
//   - публичные анализаторы третьих сторон:
//
//   - errcheck — проверяет, что ошибки не игнорируются
//
//   - ineffassign — ищет бессмысленные присваивания
//
//   - собственный анализатор osexit, запрещающий прямой os.Exit в main.main.
package main
