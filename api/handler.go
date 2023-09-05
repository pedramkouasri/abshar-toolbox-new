package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/db"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/rollback"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/update"
)

type ResponseServer struct {
	IsCompleted bool   `json:"is_complated"`
	IsFailed    bool   `json:"is_failed"`
	MessageFail string `json:"message_fail"`
	Percent     string `json:"percent"`
	State       string `json:"state"`
}

func HandleFunc(cnf config.Config, server *Server) {
	server.HandleFunc("/ping", pingHandle)
	server.HandleFunc("/patch", patchHandle(cnf))
	server.HandleFunc("/state", stateHandle)

	server.HandleFunc("/stop", stopHandle(server))
}

func stopHandle(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		server.Stop()
	}
}

func pingHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func patchHandle(cnf config.Config) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		defer w.Header().Set("Content-Type", "application/json")

		queryParams := r.URL.Query()
		version := queryParams.Get("version")

		if version == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "version is required"
			}`))
			return
		}

		fileSrc := cnf.DockerComposeDir + "/baadbaan_new/storage/app/patches/" + version
		_ = fileSrc
		// if !utils.FileExists(fileSrc) {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(`{
		// 		"message": "file not exists"
		// 	}`))
		// 	return
		// }

		db.StoreInit(version)
		logger.Info("Started")

		updateResultChan := make(chan bool)
		defer close(updateResultChan)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("Run Go Routine Update")

			up := update.NewUpdateService(cnf)

			if err := up.Handle(); err == nil {
				logger.Info("Completed Update")
				db.StoreSuccess()
				return
			}

			logger.Error(fmt.Errorf("Update Failed"))

			ctxRollback, cancelRollback := context.WithTimeout(context.Background(), cnf.RollbackTimeOut)
			defer cancelRollback()

			rollbackResultChan := make(chan bool)
			defer close(rollbackResultChan)

			rol := rollback.NewRollbackService(cnf)
			go rol.Handle(ctxRollback, rollbackResultChan)

			if res := <-rollbackResultChan; res {
				logger.Info("Rollback Success")
				return
			}
			logger.Error(fmt.Errorf("Rollback Failed"))
		}()

		wg.Wait()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": "GOOD"
		}`))
	}
}

func stateHandle(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	version := queryParams.Get("version")

	w.Header().Set("Content-Type", "application/json")

	if version == "" {
		w.WriteHeader(http.StatusBadRequest)

		err := json.NewEncoder(w).Encode(&ResponseServer{
			IsCompleted: false,
			IsFailed:    true,
			MessageFail: "version is required",
			Percent:     "0",
			State:       "",
		})

		if err != nil {
			panic(err)
		}

		return
	}

	store := db.NewBoltDB()

	// p := store.Get(fmt.Sprintf(db.Format, patchId, db.Processed))
	percent := store.Get(fmt.Sprintf(db.Format, version, db.Percent))
	isComplete := store.Get(fmt.Sprintf(db.Format, version, db.IsCompleted))
	isFailed := store.Get(fmt.Sprintf(db.Format, version, db.IsFailed))
	messageFail := store.Get(fmt.Sprintf(db.Format, version, db.MessageFail))
	state := store.Get(fmt.Sprintf(db.Format, version, db.State))

	if len(percent) == 0 {
		w.WriteHeader(http.StatusOK)
		res := &ResponseServer{
			IsCompleted: false,
			IsFailed:    false,
			MessageFail: "",
			Percent:     "0",
			State:       "Not Started",
		}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			panic(err)
		}
		return
	}

	if isFailed[0] == 1 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	res := ResponseServer{
		IsCompleted: isComplete[0] == 1,
		IsFailed:    isFailed[0] == 1,
		MessageFail: string(messageFail),
		Percent:     string(percent),
		State:       string(state),
	}

	if err := json.NewEncoder(w).Encode(&res); err != nil {
		panic(err)
	}
}
