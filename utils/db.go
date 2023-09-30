package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
)

var backupSqlDir string

func init() {
	cd, _ := os.Getwd()
	backupSqlDir = cd + "/backupSql"
}

func BackupDatabase(fileName string, dockerComposeDir string, cnf *ConfigService) error {
	err := os.Mkdir(backupSqlDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			// fmt.Println("The directory named", backupSqlDir, "exists")
		} else {
			return fmt.Errorf("Create backupSql Directory Failed error is: %v", err)
		}
	}

	sqlFileName := fmt.Sprintf("%s.sql", fileName)
	os.Remove(backupSqlDir + "/" + sqlFileName)

	file, err := os.Create(backupSqlDir + "/" + sqlFileName)
	if err != nil {
		return fmt.Errorf("Create sql fileFailed error is: %v", err)
	}
	defer file.Close()

	host, _ := cnf.Get("DB_HOST")
	port, _ := cnf.Get("DB_PORT")
	datbase, _ := cnf.Get("DB_DATABASE")
	username, _ := cnf.Get("DB_USERNAME")
	password, _ := cnf.Get("DB_PASSWORD")
	sqlCommand := fmt.Sprintf(sqlDumpCommand, strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(host), strings.TrimSpace(port), strings.TrimSpace(datbase))

	var command []string
	command = strings.Fields(fmt.Sprintf(`docker compose -f %s/docker-compose.yaml run --rm baadbaan_db %s`, dockerComposeDir, sqlCommand))

	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = file

	bufE := bytes.NewBuffer([]byte{})
	cmd.Stderr = bufE

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Dump sql Failed error is: %v %s", err, bufE.String())
	}

	return nil
}

func RestoreDatabase(fileName string, dockerComposeDir string, cnf *ConfigService) error {
	sqlFileName := fmt.Sprintf("%s.sql", fileName)
	sqlPath := backupSqlDir + "/" + sqlFileName
	if !FileExists(sqlPath) {
		return fmt.Errorf("File DB not found")
	}

	// Read the dump file
	dumpFile, err := os.Open(sqlPath)
	if err != nil {
		return fmt.Errorf("File can not open %v", err)
	}
	defer dumpFile.Close()

	host, _ := cnf.Get("DB_HOST")
	port, _ := cnf.Get("DB_PORT")
	database, _ := cnf.Get("DB_DATABASE")
	username, _ := cnf.Get("DB_USERNAME")
	password, _ := cnf.Get("DB_PASSWORD")
	sqlCommand := fmt.Sprintf(sqlRestoreDB, username, password, host, port, database)

	var command []string
	command = strings.Fields(fmt.Sprintf("docker compose -f %s/docker-compose.yaml run --rm -exec -T baadbaan_db %s", dockerComposeDir, sqlCommand))

	logger.Info(strings.Join(command, " "))

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = dumpFile

	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Restore DB sql Failed error is: %v", err)
	}

	return nil
}
