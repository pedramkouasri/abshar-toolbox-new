package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

const (
	backaupSqlDir = "./backaupSql"
)

func BackupDatabase(serviceName string, cnf *ConfigService) error {
	err := os.Mkdir(backaupSqlDir, 0755)
	if err != nil {
		if os.IsExist(err) {
			// fmt.Println("The directory named", backaupSqlDir, "exists")
		} else {
			return fmt.Errorf("Create backaupSql Directory Failed error is: %v", err)
		}
	}

	sqlFileName := fmt.Sprintf("%s-%d.sql", serviceName, current_time.Unix())
	file, err := os.Create(backaupSqlDir + "/" + sqlFileName)
	if err != nil {
		return fmt.Errorf("Create sql fileFailed error is: %v", err)
	}
	defer file.Close()

	host, _ := cnf.Get("DB_HOST")
	port, _ := cnf.Get("DB_PORT")
	datbase, _ := cnf.Get("DB_DATABASE")
	username, _ := cnf.Get("DB_USERNAME")
	password, _ := cnf.Get("DB_PASSWORD")
	sqlCommand := fmt.Sprintf(sqlDumpCommand, username, password, host, port, datbase)

	var command []string
	composeDir := viper.GetString("patch.update.docker-compose-directory") + "/docker-compose.yaml"
	command = strings.Fields(fmt.Sprintf(`docker compose -f %s run --rm %s %s`, composeDir, host, sqlCommand))

	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stdout = file
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Dump sql Failed error is: %v", err)
	}

	return nil
}
