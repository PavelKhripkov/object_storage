package v1

import (
	"encoding/json"
	file_server2 "github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server_service"
	"github.com/PavelKhripkov/object_storage/internal/domain/usecase/file_server_usecase"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type fileServerHandler struct {
	FileServerUsecase *file_server_usecase.Usecase
	l                 *log.Entry
}

func NewFileServerHandler(usecase *file_server_usecase.Usecase, l *log.Logger) Handler {
	return &fileServerHandler{
		FileServerUsecase: usecase,
		l:                 l.WithField("component", "FileServerHandler"),
	}
}

func (s fileServerHandler) Register(router *httprouter.Router) {
	router.POST("/file_server/add/:type", s.Add)
	router.GET("/file_server/:id", s.Get)
}

func (s fileServerHandler) Add(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	defer func() {
		if err := r.Body.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	var (
		err error
		res file_server2.FileServer
	)

	switch params.ByName("type") {
	case "ssh":
		var dto file_server_service.AddSSHFileServerDTO
		err = decoder.Decode(&dto)
		if err != nil {
			s.l.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err = s.FileServerUsecase.Add(r.Context(), dto)
	case "api":
		var dto file_server_service.AddAPIFileServerDTO
		err = decoder.Decode(&dto)
		if err != nil {
			s.l.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res, err = s.FileServerUsecase.Add(r.Context(), dto)

	}

	if err != nil {
		s.l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		s.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}
}

func (s fileServerHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	res, err := s.FileServerUsecase.Get(r.Context(), params.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}

	return
}
