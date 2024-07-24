package router

import (
	"baner_service/internal/cache"
	"baner_service/internal/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	redis      *cache.Cache
	db         *db.Postgres
	userToken  string
	adminToken string
}

func (h Handler) GetUsewrBanner(w http.ResponseWriter, r *http.Request) {

	//	:::: QUERY params

	tagIdStr := r.URL.Query().Get("tag_id")
	featureIdStr := r.URL.Query().Get("feature_id")
	useLastRevision := r.URL.Query().Has("use_last_revision")

	if tagIdStr == "" || featureIdStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing tag_id or feature_id+err.Error()")
		return
	}

	tagId, err := strconv.Atoi(tagIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "error parsing tag_id"+err.Error())
		return
	}
	featureId, err := strconv.Atoi(featureIdStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "error parsing feature_id"+err.Error())
		return
	}

	key := tagIdStr + featureIdStr

	//	 ::: check if it'need for last revision

	if !useLastRevision {

		gotta, err := h.redis.GetFrom(r.Context(), key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error getting from redis"+err.Error())
			return
		}
		if gotta != "" {
			w.Write([]byte(gotta))
			return
		}

	}
	//	::: GET directly from db

	bannerStr, err := h.db.GetUserBanner(r.Context(), featureId, tagId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "error getting banner"+err.Error())

		return
	}

	err = h.redis.SetTo(r.Context(), key, bannerStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "set to redis"+err.Error())

		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(bannerStr))

}

func (h Handler) GetAdminBanner(w http.ResponseWriter, r *http.Request) {

	//	::: GET params (no required)

	tagIdStr := r.URL.Query().Get("tag_id")
	featureIdStr := r.URL.Query().Get("feature_id")
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	var tagId, featureId, offset, limit int
	var err error
	if tagIdStr != "" || featureIdStr != "" || limitStr != "" || offsetStr != "" {
		tagId, err = strconv.Atoi(tagIdStr)
		featureId, err = strconv.Atoi(featureIdStr)
		offset, err = strconv.Atoi(offsetStr)
		limit, err = strconv.Atoi(limitStr)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	}

	//	::: GET from db
	banners, err := h.db.GetAdminBanner(r.Context(), tagId, featureId, offset, limit)
	log.Println("ALL BANNERS ::: ", banners)
	bannersByte, err := json.Marshal(banners)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//	::: RESPONSE with banner

	w.WriteHeader(http.StatusOK)
	w.Write(bannersByte)
}

func (h Handler) PostBanner(w http.ResponseWriter, r *http.Request) {
	// :::: DECODE request body

	var banner db.Banner
	err := json.NewDecoder(r.Body).Decode(&banner)
	if err != nil {
		println("decode" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println(banner)
	//if banner == nil {
	//	http.Error(w, errors.New("nill banner").Error(), http.StatusBadRequest)
	//	return
	//}
	//	::: create in db
	err = h.db.PostBanner(r.Context(), &banner)
	if err != nil {
		println("create in db" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//	::: response

	w.WriteHeader(http.StatusCreated)
}

func (h Handler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	var banner *db.Banner
	err := json.NewDecoder(r.Body).Decode(&banner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if banner == nil {
		http.Error(w, errors.New("nill banner").Error(), http.StatusBadRequest)
		return
	}
	//	::: create in db
	id := r.URL.Query().Get("id")
	var idInt int
	if idInt, err = strconv.Atoi(id); id == "" || err != nil {
		http.Error(w, errors.New("empty id").Error(), http.StatusBadRequest)
		return

	}
	err = h.db.UpdateBanner(r.Context(), banner, idInt, &banner.IsActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//	::: response

	w.WriteHeader(http.StatusOK)

}

func (h Handler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	var err error
	//	::: create in db
	id := r.URL.Query().Get("id")
	var idInt int
	if idInt, err = strconv.Atoi(id); id == "" || err != nil {
		http.Error(w, errors.New("empty id").Error(), http.StatusBadRequest)
		return

	}
	err = h.db.DeleteBanner(r.Context(), idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//	::: response

	w.WriteHeader(http.StatusNoContent)

}

func Register(redis *cache.Cache, db *db.Postgres, userToken, adminToken string) *chi.Mux {
	handler := &Handler{
		redis:      redis,
		db:         db,
		userToken:  userToken,
		adminToken: adminToken,
	}

	r := chi.NewRouter()
	// :::: use CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// :::: use logger
	r.Use(middleware.Logger)
	// :::: use auth
	r.Use(isAuth)

	// :::: router endpoints

	//				:::: user one
	r.Group(func(r chi.Router) {
		r.Use(guard(userToken))
		r.Get("/user-banner", handler.GetUsewrBanner)

	})

	//				:::: admin only one
	r.Group(func(r chi.Router) {
		r.Use(guard(adminToken))
		r.Get("/banner", handler.GetAdminBanner)
		r.Post("/banner", handler.PostBanner)
		r.Patch("/banner", handler.UpdateBanner)
		r.Delete("/banner", handler.DeleteBanner)

	})

	return r

}

func guard(tokens ...string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("token")
			println("TOKEN ::: ", tokens)
			for _, t := range tokens {
				if t == token {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusForbidden)
		})
	}
}

func isAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("token"); token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			handler.ServeHTTP(w, r)
		}

	})
}
