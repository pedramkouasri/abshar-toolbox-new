package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/db"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/backup"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/restore"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/rollback"
	"github.com/pedramkousari/abshar-toolbox-new/scripts/update"
	"github.com/pedramkousari/abshar-toolbox-new/types"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
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
	server.HandleFunc("/patch", patchHandle(cnf, server))
	server.HandleFunc("/state", stateHandle)

	server.HandleFunc("/backup", backupHandle(cnf, server))
	// server.HandleFunc("/state-backup", stateHandle)

	server.HandleFunc("/restore", restoreHandle(cnf, server))
	// server.HandleFunc("/state-restore", stateHandle)

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

func patchHandle(cnf config.Config, server *Server) func(w http.ResponseWriter, r *http.Request) {

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

		if !utils.FileExists(fileSrc) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "file not exists"
			}`))
			return
		}

		db.StoreInit(version)
		logger.Info("Started")

		go func() {
			diffPackages, err := StartUpdate(fileSrc, cnf)
			if err != nil {
				db.StoreError(err)
				logger.Error(fmt.Errorf("Start Failed %v", err))
				return
			}

			logger.Info("Run Go Routine Update")
			up := update.NewUpdateService(cnf)

			err = up.Handle(diffPackages)
			if err == nil {
				logger.Info("Completed Update")
				db.StoreSuccess()

				go func() {
					time.Sleep(time.Second * 12)
					server.Stop()
				}()
				return
			}

			logger.Error(fmt.Errorf("Update Failed %v", err))
			db.StoreError(err)

			rol := rollback.NewRollbackService(cnf)
			err = rol.Handle(diffPackages)
			if err == nil {
				logger.Info("Completed Rollback")
				return
			}
			logger.Error(fmt.Errorf("Rollback Failed %v", err))
		}()

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

func StartUpdate(fileSrc string, cnf config.Config) ([]types.CreatePackageParams, error) {
	if err := os.Mkdir("./temp", 0755); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("create directory err: %s", err)
		}
	}
	logger.Info("Created Temp Directory")

	if err := utils.DecryptFile([]byte(cnf.EncryptKey), fileSrc, strings.TrimSuffix(fileSrc, ".enc")); err != nil {
		return nil, fmt.Errorf("Decrypt File err: %s", err)
	}

	logger.Info("Decrypted File")

	if err := utils.UntarGzip(strings.TrimSuffix(fileSrc, ".enc"), "./temp"); err != nil {
		return nil, fmt.Errorf("UnZip File err: %s", err)
	}
	logger.Info("UnZiped File")

	packagePathFile := "./temp/package.json"

	if _, err := os.Stat(packagePathFile); err != nil {
		return nil, fmt.Errorf("package.json is err: %s", err)
	}

	logger.Info("Exists package.json")

	file, err := os.Open(packagePathFile)
	if err != nil {
		return nil, fmt.Errorf("open package.json is err: %s", err)
	}
	logger.Info("Opened package.json")

	pkg := []types.Packages{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pkg)
	if err != nil {
		return nil, fmt.Errorf("decode package.json is err: %s", err)
	}

	logger.Info("Decode package.json")

	diffPackages := utils.GetPackageDiff(pkg)
	if len(diffPackages) == 0 {
		return nil, fmt.Errorf("Not Found Diff Packages")
	}

	return diffPackages, nil
}

func backupHandle(cnf config.Config, server *Server) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		defer w.Header().Set("Content-Type", "application/json")

		branchName := time.Now().Format("2006-01-02-15-04-05")

		logger.Info("Backup Started")

		// go func() {

		logger.Info("Run Go Routine Update")
		bk := backup.NewBackupService(cnf)

		err := bk.Handle(branchName)
		if err != nil {
			logger.Error(fmt.Errorf("Backup Failed %v", err))

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "Backup Failed"
			}`))

			return
		}

		// }()
		logger.Info("Completed Backup")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{
			"file-name": "%s.tar.gz"
			"message": "Backup Completed"
		}`, branchName)))
	}
}

func restoreHandle(cnf config.Config, server *Server) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		defer w.Header().Set("Content-Type", "application/json")

		queryParams := r.URL.Query()
		filename := queryParams.Get("filename")

		if filename == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "File name is required"
			}`))
			return
		}

		fileSrc := cnf.DockerComposeDir + "/baadbaan_new/storage/app/backup/" + filename

		if !utils.FileExists(fileSrc) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "file not exists"
			}`))
			return
		}

		logger.Info("Restore Started")

		// go func() {

		logger.Info("Run Go Routine Update")
		branch, err := StartRestore(fileSrc)
		if err != nil {
			logger.Error(fmt.Errorf("Start Restore Failed %v", err))
			return
		}

		restpreFilePath := cnf.DockerComposeDir + "/baadbaan_new/storage/framework/restore"
		restoreFile, err := os.Create(restpreFilePath)
		if err != nil {
			logger.Error(fmt.Errorf("cannot create restore file in baadbaan storeage error is: %v", err))
		}
		restoreFile.Close()
		defer os.Remove(restpreFilePath)

		rs := restore.NewRestoreService(cnf)
		if err := rs.Handle(branch); err != nil {
			logger.Error(fmt.Errorf("Restore Failed %v", err))

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"message": "Restore Failed"
			}`))

			return
		}

		// }()
		logger.Info("Completed Restore")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": "Restore Completed"
		}`))
	}
}

func StartRestore(fileSrc string) (string, error) {
	if err := os.Mkdir("./temp", 0755); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("create directory err: %s", err)
		}
	}

	if err := utils.UntarGzip(fileSrc, "./temp"); err != nil {
		return "", fmt.Errorf("UnZip File err: %s", err)
	}

	branchFile := "./temp/branch.txt"

	if _, err := os.Stat(branchFile); err != nil {
		return "", fmt.Errorf("branch.txt is err: %s", err)
	}

	file, err := os.Open(branchFile)
	if err != nil {
		return "", fmt.Errorf("open branch.txt is err: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var branch string
	if scanner.Scan() {
		branch = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("cannot scan file %v", err)
	}

	return branch, nil
}
