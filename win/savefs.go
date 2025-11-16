package main

import (
	"embed"
	"io"
	"os"
	"path/filepath"
)

// copyDir рекурсивно копирует содержимое директории из fs в директорию на диске
func copyDir(fs embed.FS, srcDir, dstDir string) error {
	entries, err := fs.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.ToSlash(filepath.Join(srcDir, entry.Name()))
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Создаём директорию
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			// Рекурсивно копируем поддиректорию
			if err := copyDir(fs, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Копируем файл
			err1 := copyFile(fs, srcPath, dstPath)
			if err1 != nil {
				return err1
			}
		}
	}

	return nil
}

func copyFile(fs embed.FS, srcPath string, dstPath string) error {
	file, err := fs.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		return err
	}
	return nil
}

