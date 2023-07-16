package v1

import (
	"encoding/json"
	item_usecase "github.com/PavelKhripkov/object_storage/internal/domain/usecase/item_usecase"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
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

// Get replies with a single entity of item.
func (s itemHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	item, err := s.itemUsecase.Get(r.Context(), params.ByName("id"))
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

// Store parses body into form and passes incoming file to be stored into chunks on file servers.
func (s itemHandler) Store(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
	defer func() {
		if err := r.Body.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mr := multipart.NewReader(r.Body, params["boundary"])
	form, err := mr.ReadForm(MaxMultiPartMemory)

	cleanUpForm := func() {
		if err := form.RemoveAll(); err != nil {
			s.l.Error(err)
		}
	}

	if len(form.File) != 1 {
		defer cleanUpForm()
		http.Error(w, "accepting exactly one file", http.StatusBadRequest)
		return
	}

	if len(form.File["item"]) != 1 {
		defer cleanUpForm()
		http.Error(w, "accepting exactly one file", http.StatusBadRequest)
		return
	}
	fileHeader := form.File["item"][0]

	if len(form.Value["container_id"]) != 1 {
		defer cleanUpForm()
		http.Error(w, "accepting exactly one 'container_id' value", http.StatusBadRequest)
		return
	}
	containerID := form.Value["container_id"][0]

	dto := item_usecase.StoreItemDTO{
		F:           fileHeader,
		Name:        fileHeader.Filename,
		ContainerID: containerID,
		Size:        fileHeader.Size,
		Close:       cleanUpForm,
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

// Download replies with a stream mapped to the chunks of an item.
// Allows to start download immediately, without waiting chunks to be taken from file servers.
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
