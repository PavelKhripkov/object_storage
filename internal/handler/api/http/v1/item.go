package v1

import (
	"fmt"
	item_usecase "github.com/PavelKhripkov/object_storage/internal/domain/usecase/item"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
)

const MaxFileSize = 1 * 1024 * 1024 * 1024 // 1 Gb

type itemHandler struct {
	ItemUsecase *item_usecase.Usecase
}

func NewItemHandler(usecase *item_usecase.Usecase) Handler {
	return &itemHandler{
		ItemUsecase: usecase,
	}
}

func (s itemHandler) Register(router *httprouter.Router) {
	router.POST("/item/create", s.Create)
}

func (s itemHandler) Create(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Println("Start item creation.")

	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)

	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, "File is too big.", http.StatusBadRequest)
		return
	}

	fmt.Println("Getting file.")
	file, fileHeader, err := r.FormFile("item")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Processing file.")
	itemID, err := s.ItemUsecase.UploadItem(r.Context(), file, fileHeader.Filename, fileHeader.Size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := io.WriteString(w, itemID); err != nil {
		fmt.Println(err)
	}

}
