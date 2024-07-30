package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type Storage interface {
	GetURLByShortURI(shortURI string) (longURL string, found bool)
	SaveURL(longURL string) (shortURI string)
}

type InMemoryStorage struct {
	storage map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage: make(map[string]string),
	}
}

func (s *InMemoryStorage) GetURLByShortURI(shortURI string) (longURL string, found bool) {
	longURL, found = s.storage[shortURI]
	return longURL, found
}

func (s *InMemoryStorage) SaveURL(longURL string) (shortURI string) {
	shortURI = util.RandStringRunes(10)
	for s.storage[shortURI] != "" {
		shortURI = util.RandStringRunes(10)
	}

	s.storage[shortURI] = longURL

	return shortURI
}

type storageJSON struct {
	UUID     string `json:"uuid"`
	ShortURI string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

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
		var storageJSON storageJSON
		err = json.Unmarshal(fileScanner.Bytes(), &storageJSON)
		if err != nil {
			err = fmt.Errorf("storage: error when read json from file[%s]: %v", path, err)
			return nil, err
		}

		fileJSONStorage.storage[storageJSON.ShortURI] = storageJSON.LongURL
	}

	return fileJSONStorage, nil
}

func (s *FileJSONStorage) SaveURL(longURL string) (shortURI string) {
	shortURI = s.InMemoryStorage.SaveURL(longURL)

	file, err := os.OpenFile(s.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		println(err.Error())
	}

	defer file.Close()

	storageJSONBytes, _ := json.Marshal(&storageJSON{
		UUID:     uuid.New().String(),
		ShortURI: shortURI,
		LongURL:  longURL,
	})

	storageJSONBytes = append(storageJSONBytes, '\n')
	file.Write(storageJSONBytes)

	return shortURI
}
