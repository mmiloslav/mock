package main

import (
	"net/http"
	"os"

	"github.com/mmiloslav/mock/internal/api"
	"github.com/mmiloslav/mock/internal/app"
	"github.com/mmiloslav/mock/internal/db"
	"github.com/mmiloslav/mock/internal/mylog"
)

func main() {
	mylog.Init()
	logger := mylog.Logger.WithField("component", "main")
	logger.Info("starting mock service")

	err := db.OpenConnection(logger)
	if err != nil {
		logger.Errorf("failed to open DB connection with error [%s]", err.Error())
		os.Exit(1)
	}

	go func() {
		logger.Info("starting mock app router on port 5081...")
		err = http.ListenAndServe(":5081", app.NewRouter())
		if err != nil {
			logger.Errorf("failed to listen and serve mock app router with error [%s]", err.Error())
			os.Exit(2)
		}
	}()

	logger.Info("starting api router on port 8081...")
	err = http.ListenAndServe(":8081", api.NewRouter())
	if err != nil {
		logger.Errorf("failed to listen and serve api router with error [%s]", err.Error())
		os.Exit(3)
	}
}
