package redirect_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"
)

func TestRedirect_Success(t *testing.T) {
	urlGetterMock := mocks.NewURLGetter(t)
	urlGetterMock.On("GetURL", mock.Anything, "test_alias").
		Return("https://google.com", nil).Once()

	router := chi.NewRouter()
	router.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

	req := httptest.NewRequest(http.MethodGet, "/test_alias", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusFound, rr.Code)
	require.Equal(t, "https://google.com", rr.Header().Get("Location"))
}

func TestRedirect_NotFound(t *testing.T) {
	urlGetterMock := mocks.NewURLGetter(t)
	urlGetterMock.On("GetURL", mock.Anything, "missing").
		Return("", service.ErrURLNotFound).Once()

	router := chi.NewRouter()
	router.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRedirect_InternalServerError(t *testing.T) {
	urlGetterMock := mocks.NewURLGetter(t)
	urlGetterMock.On("GetURL", mock.Anything, "test_alias").
		Return("", errors.New("database error")).Once()

	router := chi.NewRouter()
	router.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

	req := httptest.NewRequest(http.MethodGet, "/test_alias", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
