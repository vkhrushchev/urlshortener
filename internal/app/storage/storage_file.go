package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
)

type FileJSONStorage struct {
	InMemoryStorage
	path string
}

func NewFileJSONStorage(path string) (*FileJSONStorage, error) {
	var file *os.File
	var fileInfo os.FileInfo
	var err error

	if fileInfo, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
		file, err = os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
		if err != nil {
			err = fmt.Errorf("storage: error when create and open file: %v", err)
			return nil, err
		}
	} else {
		if fileInfo.IsDir() {
			err = fmt.Errorf("storage: path[%s] is dir", path)
			return nil, err
		}
		file, err = os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			err = fmt.Errorf("storage: error when open file: %v", err)
			return nil, err
		}
	}

	defer file.Close()

	fileJSONStorage := &FileJSONStorage{
		InMemoryStorage: *NewInMemoryStorage(),
		path:            path,
	}

	// считываем json-строки из файла path
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		var shortURLEntry dto.StorageShortURLEntry
		err = json.Unmarshal(fileScanner.Bytes(), &shortURLEntry)
		if err != nil {
			err = fmt.Errorf("storage: error when read json from file[%s]: %v", path, err)
			return nil, err
		}

		fileJSONStorage.storage[shortURLEntry.ShortURI] = &shortURLEntry
	}

	return fileJSONStorage, nil
}

func (s *FileJSONStorage) SaveURL(ctx context.Context, longURL string) (*dto.StorageShortURLEntry, error) {
	shortURLEntry, err := s.InMemoryStorage.SaveURL(ctx, longURL)
	if err != nil {
		log.Errorw(
			"storage: unexpected error when save short URL to InMemeoryStorage",
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: unexpected error when save short URL to InMemeoryStorage: %v", err)

		return nil, err
	}

	file, err := os.OpenFile(s.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Errorw(
			"storage: error when open file",
			"path", s.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when open file: %v", err)

		return nil, err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Errorw(
				"storage: error when close file",
				"path", s.path,
				"error", err.Error(),
			)
		}
	}()

	storageJSONBytes, err := json.Marshal(shortURLEntry)
	if err != nil {
		log.Errorw(
			"storage: error when marshal storageJSON to JSON",
			"path", s.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when marshal storageJSON to JSON: %v", err)

		return nil, err
	}

	storageJSONBytes = append(storageJSONBytes, '\n')
	_, err = file.Write(storageJSONBytes)
	if err != nil {
		log.Errorw(
			"storage: error when write storageJSON to file",
			"path", s.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when write storageJSON to file: %v", err)

		return nil, err
	}

	return shortURLEntry, nil
}

func (s *FileJSONStorage) DeleteByShortURIs(ctx context.Context, shortURIs []string) (int, error) {
	return 0, fmt.Errorf("storage: func DeleteByShortURIs not implemented for FileJSONStorage")
}
