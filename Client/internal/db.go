package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	user     = "root"
	password = "root"
	database = "requests"
	table    = "request_metadata"
	server   = "localhost"
	port     = "9876"
	proto    = "tcp"
	driver   = "mysql"

	database_createDB    = "Client/queries/createDB.sql"
	database_createTable = "Client/queries/createTable.sql"
)

func ReadQuery(filename string) ([]byte, error) {
	fd, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ExecuteQuery(filename string, db *sql.DB) error {
	data, err := ReadQuery(filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = db.Exec(string(data))
	if err != nil {
		return err
	}
	return nil
}

func convetMapToStr(m map[string]bool) ([]byte, error) {
	mJson, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return mJson, nil
}

func (p Global_objects) InsertContentToDb(fc FileContent) error {
	criticalWordsStr, err := convetMapToStr(fc.CriticalWords)
	if err != nil {
		p.Logger.Errorf("Could not convert %v map to string", fc.CriticalWords)
	}
	_, err = p.DBobject.Exec("INSERT INTO `data` (`Date`, `Session`, `IP`, `IpType`, `UA`, `Country`, `Path`, `Method`, `SessionKey`, `CriticalWords`, `Crawler`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fc.Date, fc.Session, fc.IP, fc.IpType, fc.UA, fc.Country, fc.Path, fc.Method, fc.SessionKey, string(criticalWordsStr), strconv.FormatBool(fc.Crawler))
	if err != nil {
		return err
	}
	return nil
}

func (p Global_objects) InsertMetadataToDb(filename, fileHash, pullTime, lastModified, etag string, size int64) error {
	_, err := p.DBobject.Exec("INSERT INTO `metadata` (`pull_time`, `updated_time`, `filename`, `file_size`, `file_hash`, `Etag`) VALUES (?, ?, ?, ?, ?, ?)", pullTime, lastModified, filename, size, fileHash, etag)
	if err != nil {
		return err
	}
	return nil
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", user+":"+password+"@"+proto+"("+server+":"+port+")"+"/")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if err = ExecuteQuery(database_createDB, db); err != nil {
		fmt.Println(err)
		return nil, err
	}
	db.Close()
	time.Sleep(time.Second * 1)
	db, err = sql.Open(driver, user+":"+password+"@"+proto+"("+server+":"+port+")"+"/"+database+"?multiStatements=true")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if err = ExecuteQuery(database_createTable, db); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return db, nil

}
