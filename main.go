package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Текстовый анализатор ===")
	fmt.Println("Доступные команды:")
	fmt.Println("  анализ <путь_к_файлу> <фраза_для_поиска> - Проанализировать файл")
	fmt.Println("  выход - Выйти из программы")
	fmt.Println("==============================")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		if input == "выход" {
			fmt.Println("Программа завершена.")
			break
		}

		parts := strings.Fields(input)
		if len(parts) < 3 || parts[0] != "анализ" {
			fmt.Println("Ошибка: неверная команда. Используйте 'анализ <путь_к_файлу> <фраза_для_поиска>'")
			continue
		}

		filePath := parts[1]
		phraseToSearch := strings.Join(parts[2:], " ")

		content, err := readFile(filePath)
		if err != nil {
			fmt.Println("Ошибка чтения файла:", err)
			continue
		}

		totalWords, phraseCount := searchPhrase(content, phraseToSearch)

		fmt.Printf("Общее количество слов в файле: %d\n", totalWords)
		fmt.Printf("Количество вхождений фразы '%s': %d\n", phraseToSearch, phraseCount)
	}
}

func readFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + " ")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content.String(), nil
}

func searchPhrase(content, phrase string) (int, int) {
	words := strings.Fields(content)
	totalWords := len(words)
	phraseCount := strings.Count(content, phrase)

	return totalWords, phraseCount
}
