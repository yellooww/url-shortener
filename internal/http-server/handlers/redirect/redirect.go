package redirect

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLGetter --dir=. --output=./mocks --outpkg=mocks

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.redirect.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		url, err := urlGetter.GetURL(r.Context(), alias)

		var validErrs validator.ValidationErrors
		if errors.As(err, &validErrs) {
			log.Error("invalid request", "alias", alias, "error", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validErrs))
			return
		}
		if errors.Is(err, service.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", "alias", alias)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get url"))
			return
		}

		log.Info("got url", slog.String("url", url))
		http.Redirect(w, r, url, http.StatusFound)
	}
}
