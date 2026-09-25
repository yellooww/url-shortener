package delete_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDelete_Success(t *testing.T) {
	urlDeleterMock := mocks.NewURLDeleter(t)
	urlDeleterMock.On("DeleteURL", mock.Anything, "test_alias").
		Return(int64(1), nil).Once()

	router := chi.NewRouter()
	router.Delete("/{alias}", delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

	req := httptest.NewRequest(http.MethodDelete, "/test_alias", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"countDeleted":1`)
}

func TestDelete_NotFound(t *testing.T) {
	urlDeleterMock := mocks.NewURLDeleter(t)
	urlDeleterMock.On("DeleteURL", mock.Anything, "missing").
		Return(int64(0), service.ErrURLNotFound).Once()

	router := chi.NewRouter()
	router.Delete("/{alias}", delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

	req := httptest.NewRequest(http.MethodDelete, "/missing", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestDelete_InternalServerError(t *testing.T) {
	urlDeleterMock := mocks.NewURLDeleter(t)
	urlDeleterMock.On("DeleteURL", mock.Anything, "test_alias").
		Return(int64(0), errors.New("database error")).Once()

	router := chi.NewRouter()
	router.Delete("/{alias}", delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

	req := httptest.NewRequest(http.MethodDelete, "/test_alias", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
