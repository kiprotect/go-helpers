// KIProtect Go-Helpers - Golang Utility Functions
// Copyright (C) 2020  KIProtect GmbH (HRB 208395B) - Germany
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package migrate

import (
	"database/sql"
	"fmt"
	"github.com/kiprotect/go-helpers/interpolate"
	"github.com/kiprotect/go-helpers/maps"
	"github.com/kiprotect/go-helpers/yaml"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type ConfigError struct {
	msg string
}

type Migration struct {
	Version     int
	Description string
	Content     string
	FileName    string
}

func (self *ConfigError) Error() string {
	return self.msg
}

type MigrationManager struct {
	Config         map[string]interface{}
	Path           string
	UpMigrations   map[int]Migration
	DownMigrations map[int]Migration
	latestVersion  int
	DB             *sql.DB
}

func (self *MigrationManager) LatestVersion() int {
	return self.latestVersion
}

func (self *MigrationManager) CurrentVersion() (int, error) {

	//we test the SQL connection
	_, err := self.DB.Query(`
  	      SELECT 'Hello, World';
	    `)

	if err != nil {
		return 0, err
	}

	interpolatedString, err := interpolate.Interpolate(`
        SELECT
            {.version_table.version_column}
        FROM
            {.version_table.name}
        LIMIT 1;
    `, self.Config)

	if err != nil {
		//the version table does not exist yet...
		return 0, nil
	}
	rows, err := self.DB.Query(interpolatedString)

	if err != nil {
		return 0, nil
	}

	if !rows.Next() {
		//nothing stored in the version table so far
		return 0, nil
	}

	var version int
	rows.Scan(&version)

	return version, nil
}

//Migrates the database to the current head version
func (self *MigrationManager) Migrate(version int) error {
	currentVersion, _ := self.CurrentVersion()
	relevantMigrations := make([]Migration, 0, 10)
	up := true
	if version != -1 && version < currentVersion {
		up = false
	}
	if up {
		keys := make([]int, 0)
		for key := range self.UpMigrations {
			if key > currentVersion {
				// if an explicit version number is given,
				// we include only revisions up to this version (inclusive)
				if version == -1 || key <= version {
					keys = append(keys, key)
				}
			}
		}
		sort.Ints(keys)
		for _, k := range keys {
			relevantMigrations = append(relevantMigrations, self.UpMigrations[k])
		}
	} else {
		keys := make([]int, 0)
		for key := range self.DownMigrations {
			if key <= currentVersion {
				// if an explicit version number is given,
				// we include only revisions up to this version (non-inclusive)
				if version == -1 || key > version {
					keys = append(keys, key)
				}
			}
		}
		sort.Sort(sort.Reverse(sort.IntSlice(keys)))
		for _, k := range keys {
			relevantMigrations = append(relevantMigrations, self.DownMigrations[k])
		}
	}
	tx, err := self.DB.Begin()
	if err != nil {
		return err
	}
	err = self.ExecuteMigrations(relevantMigrations)

	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return err
}

//Executes a list of migrations
func (self *MigrationManager) ExecuteMigrations(migrations []Migration) error {
	for _, migration := range migrations {
		log.Printf("Executing migration %v\n", migration.FileName)
		_, err := self.DB.Exec(migration.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

//Load the migrations from the "migrations" subfolder
func (self *MigrationManager) LoadMigrations() error {
	self.latestVersion = 0
	fileInfos, err := ioutil.ReadDir(self.Path)
	if err != nil {
		return err
	}
	re := regexp.MustCompile("(?i)^(\\d+)_(up|down)_(.*)\\.sql$")
	for _, fileInfo := range fileInfos {
		subMatches := re.FindStringSubmatch(fileInfo.Name())
		if subMatches == nil {
			continue
		}
		version, _ := strconv.Atoi(subMatches[1])
		if version > self.latestVersion {
			self.latestVersion = version
		}
		direction := subMatches[2]
		description := subMatches[3]
		migrationFileName := filepath.Join(self.Path, fileInfo.Name())
		content, err := ioutil.ReadFile(migrationFileName)
		if err != nil {
			continue
		}
		migration := Migration{
			Description: description,
			Content:     string(content),
			Version:     version,
			FileName:    migrationFileName,
		}
		if direction == "up" {
			self.UpMigrations[version] = migration
		} else {
			self.DownMigrations[version] = migration
		}
	}
	return nil
}

func (self *MigrationManager) LoadConfig(ConfigPath string) error {

	var config map[string]interface{}
	filePath := path.Join(ConfigPath, "config.yml")
	fileContent, err := ioutil.ReadFile(filePath)

	if err != nil {
		return err
	}

	yamlerror := yaml.Unmarshal(fileContent, &config)

	if yamlerror != nil {
		return fmt.Errorf("unmarshal %v: %v", filePath, yamlerror)
	}

	deepStringConfig, ok := maps.EnsureStringKeys(config)

	if !ok {
		return fmt.Errorf("Non-string keys encountered in file '%s'", filePath)
	}

	self.Config = deepStringConfig.(map[string]interface{})

	return nil
}

func MakeMigrationManager(configPath string, db *sql.DB) (*MigrationManager, error) {
	migrationManager := new(MigrationManager)
	migrationManager.Path = configPath
	migrationManager.DB = db
	err := migrationManager.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("migrationManager.LoadConfig: %v", err)
	}
	migrationManager.UpMigrations = make(map[int]Migration)
	migrationManager.DownMigrations = make(map[int]Migration)
	err = migrationManager.LoadMigrations()
	if err != nil {
		return nil, err
	}
	return migrationManager, nil
}
