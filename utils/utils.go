package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var currentDirectory string

func init() {
	currentDirectory, _ = os.Getwd()

	os.RemoveAll(currentDirectory + "/temp")

	err := os.Mkdir(currentDirectory+"/temp", 0755)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("The directory named", currentDirectory+"/temp", "exists")
		} else {
			log.Fatalln(err)
		}
	}

	os.Mkdir(currentDirectory+"/temp/builds", 0755)
}

/*
*
username := "www-data"
chown -R www-data:www-data /path/to/your/laravel/root/directory
sudo find /path/to/your/laravel/root/directory -type f -exec chmod 644 {} \;
sudo find /path/to/your/laravel/root/directory -type d -exec chmod 755 {} \;
sudo chgrp -R www-data storage bootstrap/cache
sudo chmod -R ug+rwx storage bootstrap/cache
*/
func ChangePermision(username string, dir string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("Error retrieving information for user %s: %s\n", username, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("Error retrieving convert uid to string for uid %s: %s\n", u.Uid, err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("Error retrieving convert gid to string for gid %s: %s\n", u.Gid, err)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fi, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if err = os.Chown(path, uid, gid); err != nil {
			return fmt.Errorf("Failed to change ownership of %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk to filepath error in err  %s\n", err)
	}

	return nil
}

func getCommand(cmd string, containerName string) []string {
	return strings.Fields(fmt.Sprintf("docker exec %s %s", containerName, cmd))
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func RestartService(containerName string, dockerComposeDir string) error {
	var command []string
	composeDir := dockerComposeDir + "/docker-compose.yaml"
	command = strings.Fields(fmt.Sprintf(`docker compose -f %s restart %s`, composeDir, containerName))

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
