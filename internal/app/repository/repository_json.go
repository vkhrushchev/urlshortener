package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/vkhrushchev/urlshortener/internal/app/entity"
)

// JSONFileShortURLRepository реализует интерфейс IShortURLRepository для хранения коротких ссылок в json-файле
type JSONFileShortURLRepository struct {
	InMemoryShortURLRepository
	path string
}

// NewJSONFileShortURLRepository создает экземпляр структуры JSONFileShortURLRepository
func NewJSONFileShortURLRepository(path string) (*JSONFileShortURLRepository, error) {
	var file *os.File
	var fileInfo os.FileInfo
	var err error

	if fileInfo, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
		file, err = os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
		if err != nil {
			err = fmt.Errorf("repository: error when create and open file: %v", err)
			return nil, err
		}
	} else {
		if fileInfo.IsDir() {
			err = fmt.Errorf("repository: path[%s] is dir", path)
			return nil, err
		}
		file, err = os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			err = fmt.Errorf("repository: error when open file: %v", err)
			return nil, err
		}
	}

	defer func(file *os.File) {
		if fileCloseErr := file.Close(); fileCloseErr != nil {
			log.Errorw("repository: error when close file", "fileCloseErr", fileCloseErr)
		}
	}(file)

	jsonFileShortURLRepository := &JSONFileShortURLRepository{
		InMemoryShortURLRepository: *NewInMemoryShortURLRepository(),
		path:                       path,
	}

	// считываем json-строки из файла path
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		var shortURLEntity entity.ShortURLEntity
		err = json.Unmarshal(fileScanner.Bytes(), &shortURLEntity)
		if err != nil {
			err = fmt.Errorf("repository: error when read json from file[%s]: %v", path, err)
			return nil, err
		}

		jsonFileShortURLRepository.storage[shortURLEntity.ShortURI] = &shortURLEntity
	}

	return jsonFileShortURLRepository, nil
}

// SaveShortURL сохраняет короткую ссылку
func (r *JSONFileShortURLRepository) SaveShortURL(ctx context.Context, shortURLEntity *entity.ShortURLEntity) (*entity.ShortURLEntity, error) {
	shortURLEntity, err := r.InMemoryShortURLRepository.SaveShortURL(ctx, shortURLEntity)
	if err != nil {
		log.Errorw("repository: error when save short url", "err", err)
		return nil, err
	}

	file, err := os.OpenFile(r.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Errorw("repository: error when open file", "path", r.path, "err", err)
		return nil, ErrUnexpected
	}

	defer func(file *os.File) {
		if fileCloseErr := file.Close(); fileCloseErr != nil {
			log.Errorw("repository: error when close file", "fileCloseErr", fileCloseErr)
		}
	}(file)

	shortURLEntityJSONBytes, err := json.Marshal(shortURLEntity)
	if err != nil {
		log.Errorw(
			"repository: error when marshal shortURLEntity to JSON",
			"path", r.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when marshal storageJSON to JSON: %v", err)

		return nil, err
	}

	shortURLEntityJSONBytes = append(shortURLEntityJSONBytes, '\n')
	_, err = file.Write(shortURLEntityJSONBytes)
	if err != nil {
		log.Errorw(
			"storage: error when write storageJSON to file",
			"path", r.path,
			"error", err.Error(),
		)
		err = fmt.Errorf("storage: error when write storageJSON to file: %v", err)

		return nil, err
	}

	return shortURLEntity, nil
}
