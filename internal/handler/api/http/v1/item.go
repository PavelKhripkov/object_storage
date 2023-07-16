package v1

import (
	"encoding/json"
	item_usecase "github.com/PavelKhripkov/object_storage/internal/domain/usecase/item_usecase"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
)

const MaxFileSize = 10 * 1024 * 1024 * 1024  // 10 Gb
const MaxMultiPartMemory = 100 * 1024 * 1024 // 100 Mb

type itemHandler struct {
	itemUsecase *item_usecase.Usecase
	l           *log.Entry
}

func NewItemHandler(usecase *item_usecase.Usecase, l *log.Logger) Handler {
	return &itemHandler{
		itemUsecase: usecase,
		l:           l.WithField("component", "ItemHandler"),
	}
}

func (s itemHandler) Register(router *httprouter.Router) {
	router.POST("/item/store", s.Store)
	router.GET("/item/:id", s.Get)
	router.GET("/item/:id/download", s.Download)
}

func (s itemHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	item, err := s.itemUsecase.Get(r.Context(), params.ByName("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}

	return
}

func (s itemHandler) Store(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
	defer func() {
		if err := r.Body.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	if err := r.ParseMultipartForm(MaxMultiPartMemory); err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, fileHeader, err := r.FormFile("item")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	containerID := r.FormValue("container_id")

	// TODO read file directly into specified place.
	closeFunc := func() error {
		if err := f.Close(); err != nil {
			return err
		}
		if err := r.MultipartForm.RemoveAll(); err != nil {
			return err
		}

		return nil
	}

	dto := item_usecase.StoreItemDTO{
		F:           fileHeader,
		Name:        fileHeader.Filename,
		ContainerID: containerID,
		Size:        fileHeader.Size,
		Close:       closeFunc,
	}

	item, err := s.itemUsecase.Store(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = io.WriteString(w, string(bytes)); err != nil {
		s.l.Error(err)
	}

	return
}

func (s itemHandler) Download(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	contentMapper, origFileName, err := s.itemUsecase.Download(r.Context(), params.ByName("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(origFileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, origFileName, time.Time{}, contentMapper)
}
