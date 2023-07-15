package main

import (
	"context"
	sqlite2 "github.com/PavelKhripkov/object_storage/internal/adapter/db/sqlite"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/chunk"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/item_split"
	file_server_usecase "github.com/PavelKhripkov/object_storage/internal/domain/usecase/file_server"
	item_usecase "github.com/PavelKhripkov/object_storage/internal/domain/usecase/item"
	"github.com/PavelKhripkov/object_storage/internal/handler/api/http/v1"
	"github.com/PavelKhripkov/object_storage/pkg/client/sqlite"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	logger.SetReportCaller(true)

	l := logger.WithField("component", "main")

	l.Info("Starting app")
	router := httprouter.New()

	db, err := sqlite.NewClient(context.TODO(), "", "", "./object_storage.db")
	if err != nil {
		l.WithError(err).Fatal("Couldn't connect to database.")
	}
	defer db.Close()

	itemStorage := sqlite2.NewItemStorage(db, logger)
	itemService := item.NewItemService(itemStorage, logger)

	fileServerStorage := sqlite2.NewFileServerStorage(db, logger)
	fileServerService := file_server.NewFileServerService(fileServerStorage, logger)

	chunkStorage := sqlite2.NewChunkStorage(db, logger)
	chunkService := chunk.NewChunkService(chunkStorage, logger)

	splitFileService := item_split.NewFileSplitService(logger)

	itemUsecase := item_usecase.NewItemUsecase(itemService, chunkService, fileServerService, splitFileService, logger)
	itemHandler := v1.NewItemHandler(itemUsecase, logger)

	fileServerUsecase := file_server_usecase.NewFileServerUsecase(fileServerService, logger)
	fileServerHandler := v1.NewFileServerHandler(fileServerUsecase, logger)

	l.Info("Registering handlers")
	itemHandler.Register(router)
	fileServerHandler.Register(router)

	l.Info("Listening on port 11111")
	if err := http.ListenAndServe(":11111", router); err != nil {
		l.Fatal(err)
	}
}
