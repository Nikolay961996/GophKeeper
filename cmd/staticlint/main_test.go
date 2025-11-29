package main

import (
	"testing"
)

func TestMainFunction(_ *testing.T) {
	// Простейший тест - проверяем что main существует
	// Мы не можем вызвать main(), но можем проверить что пакет компилируется
}

func TestAnalyzerFunctions(_ *testing.T) {
	// Проверяем что функции анализатора существуют
	_ = isProjectFile
	_ = isGeneratedFile
	_ = wrapAnalyzer

	// Эти функции используются в main, но мы не будем их вызывать
	// Просто проверяем что они объявлены
}
