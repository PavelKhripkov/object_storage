package v1

import (
	"encoding/json"
	item_usecase "github.com/PavelKhripkov/object_storage/internal/domain/usecase/item"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
)

const MaxFileSize = 10 * 1024 * 1024 * 1024  // 1 Gb
const MaxMultiPartMemory = 100 * 1024 * 1024 // 100 Mb

type itemHandler struct {
	ItemUsecase *item_usecase.Usecase
	l           *log.Entry
}

func NewItemHandler(usecase *item_usecase.Usecase, l *log.Logger) Handler {
	return &itemHandler{
		ItemUsecase: usecase,
		l:           l.WithField("component", "ItemHandler"),
	}
}

func (s itemHandler) Register(router *httprouter.Router) {
	router.POST("/item/store", s.Store)
	router.GET("/item/:id", s.Get)
	router.GET("/item/:id/download", s.Download)
}

func (s itemHandler) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	res, err := s.ItemUsecase.Get(r.Context(), params.ByName("id"))
	if err != nil {
		// TODO may face different errors
		w.WriteHeader(http.StatusInternalServerError)
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

//func (s itemHandler) Store(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
//	defer r.Body.Close()
//
//	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
//	if err != nil {
//		s.l.Fatal(err)
//	}
//
//	s.l.Info(r.Header.Get("Content-Length"))
//
//	if strings.HasPrefix(mediaType, "multipart/") {
//		mr := multipart.NewReader(r.Body, params["boundary"])
//		for {
//			p, err := mr.NextPart()
//			if err == io.EOF {
//				s.l.Info("EOF")
//				return
//			}
//			if err != nil {
//				s.l.Fatal(err)
//			}
//			slurp, err := io.ReadAll(p)
//			if err != nil {
//				s.l.Fatal(err)
//			}
//			s.l.Info(len(slurp))
//			//fmt.Printf("Part %q: %q\n", p.Header.Get("Foo"), slurp)
//		}
//	}
//}

func (s itemHandler) Store(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
	defer r.Body.Close()

	if err := r.ParseMultipartForm(MaxMultiPartMemory); err != nil {
		s.l.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, fileHeader, err := r.FormFile("item")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dto := item_usecase.StoreItemDTO{
		F:        fileHeader,
		Name:     fileHeader.Filename,
		Path:     "",
		BucketID: "",
		Size:     fileHeader.Size,
	}

	item, err := s.ItemUsecase.Store(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO return item
	if _, err := io.WriteString(w, item.ID); err != nil {
		log.Println(err)
	}

}

func (s itemHandler) Download(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	contentMapper, origFileName, err := s.ItemUsecase.Download(r.Context(), params.ByName("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(origFileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, origFileName, time.Time{}, contentMapper)
}
