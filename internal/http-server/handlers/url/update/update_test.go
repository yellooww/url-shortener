package update_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/update"
	"url-shortener/internal/http-server/handlers/url/update/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"

	"github.com/stretchr/testify/require"
)

func TestUpdate_Success(t *testing.T) {
	urlUpdaterMock := mocks.NewURLUpdater(t)
	urlUpdaterMock.On("UpdateURL", "test_alias", "https://google.com").
		Return(int64(1), nil).Once()

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	req := httptest.NewRequest(
		http.MethodPut,
		"/update",
		bytes.NewBufferString(`{"alias":"test_alias","url":"https://google.com"}`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"status":"OK"`)
	require.Contains(t, rr.Body.String(), `"countUpdated":1`)
}

func TestUpdate_BadRequest(t *testing.T) {
	urlUpdaterMock := mocks.NewURLUpdater(t)

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	req := httptest.NewRequest(
		http.MethodPut,
		"/update",
		bytes.NewBufferString(`{"alias":"test_alias","url":`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	urlUpdaterMock.AssertNotCalled(t, "UpdateURL")
}

func TestUpdate_NotFound(t *testing.T) {
	urlUpdaterMock := mocks.NewURLUpdater(t)
	urlUpdaterMock.On("UpdateURL", "missing", "https://google.com").
		Return(int64(0), service.ErrURLNotFound).Once()

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	req := httptest.NewRequest(
		http.MethodPut,
		"/update",
		bytes.NewBufferString(`{"alias":"missing","url":"https://google.com"}`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	require.Contains(t, rr.Body.String(), `"url not found"`)
}

func TestUpdate_InternalServerError(t *testing.T) {
	urlUpdaterMock := mocks.NewURLUpdater(t)
	urlUpdaterMock.On("UpdateURL", "test_alias", "https://google.com").
		Return(int64(0), errors.New("database error")).Once()

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	req := httptest.NewRequest(
		http.MethodPut,
		"/update",
		bytes.NewBufferString(`{"alias":"test_alias","url":"https://google.com"}`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Contains(t, rr.Body.String(), `"failed to update url"`)
}
