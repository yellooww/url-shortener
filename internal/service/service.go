package service

import (
	"context"
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
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
	GetURL(ctx context.Context, alias string) (string, error)
	DeleteURL(ctx context.Context, alias string) (int64, error)
	UpdateURL(ctx context.Context, alias string, newURL string) (int64, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{storage: storage}
}

type saveRequest struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" validate:"omitempty,min=3,max=20,alphanum"`
}

type updateRequest struct {
	Alias  string `json:"alias" validate:"required,min=3,max=20,alphanum"`
	NewURL string `json:"url" validate:"required,url"`
}

type aliasRequest struct {
	Alias string `validate:"required,min=3,max=20,alphanum"`
}

func (s *Service) SaveURL(ctx context.Context, urlToSave string, alias string) (string, error) {
	if err := validator.New().Struct(saveRequest{URL: urlToSave, Alias: alias}); err != nil {
		return "", err
	}

	if alias != "" {
		_, err := s.storage.SaveURL(ctx, urlToSave, alias)
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

		_, err := s.storage.SaveURL(ctx, urlToSave, generatedAlias)
		if errors.Is(err, storage.ErrURLExists) {
			continue
		}
		if err != nil {
			return "", err
		}

		return generatedAlias, nil
	}
}

func (s *Service) GetURL(ctx context.Context, alias string) (string, error) {
	if err := validator.New().Struct(aliasRequest{Alias: alias}); err != nil {
		return "", err
	}

	url, err := s.storage.GetURL(ctx, alias)
	if errors.Is(err, storage.ErrURLNotFound) {
		return "", ErrURLNotFound
	}

	return url, err
}

func (s *Service) DeleteURL(ctx context.Context, alias string) (int64, error) {
	if err := validator.New().Struct(aliasRequest{Alias: alias}); err != nil {
		return 0, err
	}

	countDeleted, err := s.storage.DeleteURL(ctx, alias)
	if err != nil {
		return 0, err
	}
	if countDeleted == 0 {
		return 0, ErrURLNotFound
	}

	return countDeleted, nil
}

func (s *Service) UpdateURL(ctx context.Context, alias string, newURL string) (int64, error) {
	if err := validator.New().Struct(updateRequest{Alias: alias, NewURL: newURL}); err != nil {
		return 0, err
	}

	countUpdated, err := s.storage.UpdateURL(ctx, alias, newURL)
	if err != nil {
		return 0, err
	}
	if countUpdated == 0 {
		return 0, ErrURLNotFound
	}

	return countUpdated, nil
}
