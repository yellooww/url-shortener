package service

import (
	"errors"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-playground/validator"
)

const aliasLength = 5

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)

type Storage interface {
	SaveURL(urlToSave string, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) (int64, error)
	UpdateURL(alias string, newURL string) (int64, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{storage: storage}
}

type saveRequest struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type updateRequest struct {
	Alias  string `json:"alias" validate:"required"`
	NewURL string `json:"url" validate:"required,url"`
}

type aliasRequest struct {
	Alias string `validate:"required"`
}

func (s *Service) SaveURL(urlToSave string, alias string) (string, error) {
	if err := validator.New().Struct(saveRequest{URL: urlToSave}); err != nil {
		return "", err
	}

	if alias != "" {
		if err := s.aliasExists(alias); err != nil {
			return "", err
		}

		_, err := s.storage.SaveURL(urlToSave, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				return "", ErrURLExists
			}
			return "", err
		}

		return alias, nil
	}

	for {
		generatedAlias := random.NewRandomString(aliasLength)

		err := s.aliasExists(generatedAlias)
		if errors.Is(err, ErrURLExists) {
			continue
		}
		if err != nil {
			return "", err
		}

		_, err = s.storage.SaveURL(urlToSave, generatedAlias)
		if errors.Is(err, storage.ErrURLExists) {
			continue
		}
		if err != nil {
			return "", err
		}

		return generatedAlias, nil
	}
}

func (s *Service) GetURL(alias string) (string, error) {
	if err := validator.New().Struct(aliasRequest{Alias: alias}); err != nil {
		return "", err
	}

	url, err := s.storage.GetURL(alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		return "", ErrURLNotFound
	}

	return url, err
}

func (s *Service) DeleteURL(alias string) (int64, error) {
	if err := validator.New().Struct(aliasRequest{Alias: alias}); err != nil {
		return 0, err
	}

	countDeleted, err := s.storage.DeleteURL(alias)
	if err != nil {
		return 0, err
	}
	if countDeleted == 0 {
		return 0, ErrURLNotFound
	}

	return countDeleted, nil
}

func (s *Service) UpdateURL(alias string, newURL string) (int64, error) {
	if err := validator.New().Struct(updateRequest{Alias: alias, NewURL: newURL}); err != nil {
		return 0, err
	}

	countUpdated, err := s.storage.UpdateURL(alias, newURL)
	if err != nil {
		return 0, err
	}
	if countUpdated == 0 {
		return 0, ErrURLNotFound
	}

	return countUpdated, nil
}

func (s *Service) aliasExists(alias string) error {
	_, err := s.storage.GetURL(alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		return nil
	}
	return err
}
