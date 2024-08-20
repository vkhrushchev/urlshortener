package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
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

		fileJSONStorage.storage[shortURLEntry.ShortURI] = shortURLEntry.LongURL
	}

	return fileJSONStorage, nil
}

func (s *FileJSONStorage) SaveURL(ctx context.Context, longURL string) (shortURI string, err error) {
	shortURI, err = s.InMemoryStorage.SaveURL(ctx, longURL)
	if err != nil {
		log.Errorw(
			"storage: unexpected error when save short URL to InMemeoryStorage",
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: unexpected error when save short URL to InMemeoryStorage: %v", err)

		return "", err
	}

	file, err := os.OpenFile(s.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Errorw(
			"storage: error when open file",
			"path", s.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when open file: %v", err)

		return "", err
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

	storageJSONBytes, err := json.Marshal(&dto.StorageShortURLEntry{
		UUID:     uuid.New().String(),
		ShortURI: shortURI,
		LongURL:  longURL,
	})
	if err != nil {
		log.Errorw(
			"storage: error when marshal storageJSON to JSON",
			"path", s.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when marshal storageJSON to JSON: %v", err)

		return "", err
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

		return "", err
	}

	return shortURI, nil
}

func (s *FileJSONStorage) SaveURLBatch(ctx context.Context, entries []*dto.StorageShortURLEntry) ([]*dto.StorageShortURLEntry, error) {
	for _, entry := range entries {
		shortURI, err := s.SaveURL(ctx, entry.LongURL)
		if err != nil {
			return nil, err
		}

		entry.ShortURI = shortURI
	}

	return entries, nil
}

func (s *FileJSONStorage) GetURLByUserID(ctx context.Context, userID string) ([]*dto.StorageShortURLEntry, error) {
	// not supported
	return make([]*dto.StorageShortURLEntry, 0), nil
}
