package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

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
