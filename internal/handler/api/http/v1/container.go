package v1

import (
	"encoding/json"
	"github.com/PavelKhripkov/object_storage/internal/domain/usecase/container_usecase"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type containerHandler struct {
	containerUsecase *container_usecase.Usecase
	l                *log.Entry
}

func NewContainerHandler(usecase *container_usecase.Usecase, l *log.Logger) Handler {
	return &containerHandler{
		containerUsecase: usecase,
		l:                l.WithField("component", "ContainerHandler"),
	}
}

func (s containerHandler) Register(router *httprouter.Router) {
	router.POST("/container/create", s.Create)
	router.GET("/container/:id", s.Get)
	router.GET("/container", s.List)
	router.DELETE("/container/:id/delete", s.Delete)
}

// Get replies with a single entity of container.
func (s containerHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	res, err := s.containerUsecase.Get(r.Context(), params.ByName("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}

	return
}

// Create creates an item container.
func (s containerHandler) Create(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	defer func() {
		if err := r.Body.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	var dto container_usecase.CreateContainerDTO

	err := decoder.Decode(&dto)
	if err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := s.containerUsecase.Create(r.Context(), dto)

	if err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}
}

// List returns all containers.
func (s containerHandler) List(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	res, err := s.containerUsecase.List(r.Context())

	if err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}
}

func (s containerHandler) Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// TODO implement.
}
