package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"
)

func TestSave_Success(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	urlSaverMock.On("SaveURL", "https://google.com", "test_alias").
		Return("test_alias", nil).Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(
		http.MethodPost,
		"/url",
		bytes.NewBufferString(`{"url":"https://google.com","alias":"test_alias"}`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var response save.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

	require.Equal(t, "OK", response.Status)
	require.Equal(t, "test_alias", response.Alias)
}

func TestSave_BadRequest(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(
		http.MethodPost,
		"/url",
		bytes.NewBufferString(`{"url":"https://google.com"`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	require.Equal(t, "failed to decode request", response["error"])
}

func TestSave_Conflict(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	urlSaverMock.On("SaveURL", "https://google.com", "test_alias").
		Return("", service.ErrURLExists).Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	req := httptest.NewRequest(
		http.MethodPost,
		"/url",
		bytes.NewBufferString(`{"url":"https://google.com","alias":"test_alias"}`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)

	var response map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	require.Equal(t, "url already exists", response["error"])
}

func TestSave_InternalServerError(t *testing.T) {
	urlSaverMock := mocks.NewURLSaver(t)
	urlSaverMock.On("SaveURL", "https://google.com", "test_alias").
		Return("", errors.New("database error")).Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)
	req := httptest.NewRequest(
		http.MethodPost,
		"/save",
		bytes.NewBufferString(`{"url":"https://google.com","alias":"test_alias"}`),
	)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Contains(t, rr.Body.String(), `"failed to add url"`)
}
