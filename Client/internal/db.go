package internal

import (
	"crawlerDetection/Client/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

func readQuery(filename string) ([]byte, error) {
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

func executeQuery(filename string, db *sql.DB) error {
	data, err := readQuery(filename)
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
		p.Logger.Errorf("Could not convert map: %v to string", fc.CriticalWords)
	}
	_, err = p.DBobject.Exec("INSERT INTO `data` (`Datetime`, `Session`, `IP`, `IpType`, `UA`, `Country`, `Path`, `Method`, `SessionKey`, `CriticalWords`, `Crawler`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fc.DateTime, fc.Session, fc.IP, fc.IpType, fc.UA, fc.Country, fc.Path, fc.Method, fc.SessionKey, string(criticalWordsStr), strconv.FormatBool(fc.Crawler))
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

func InitDB(config utils.Config) (*sql.DB, error) {
	db, err := sql.Open(utils.Driver, config.Database.User+":"+config.Database.Password+"@"+utils.Proto+"("+utils.Server+":"+config.Database.Port+")"+"/")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if err = executeQuery(utils.Database_createDB, db); err != nil {
		fmt.Println(err)
		return nil, err
	}
	db.Close()
	time.Sleep(time.Second * 1)
	db, err = sql.Open(utils.Driver, config.Database.User+":"+config.Database.Password+"@"+utils.Proto+"("+utils.Server+":"+config.Database.Port+")"+"/"+utils.Database+"?multiStatements=true")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if err = executeQuery(utils.Database_createTable, db); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return db, nil

}
