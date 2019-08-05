package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const (
	MAX_WITh_IN     = 1000 // The largest number with in
	MAX_WITH_INSERT = 1000 // The largest number with insert value
)

var Db *gorm.DB

func Initialize(addr, user, password, dbName string, maxOpenConns, maxIdleConns, maxWaitTimeout int) error {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, password, addr, dbName)
	gdb, err := gorm.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	gdb.DB().SetMaxOpenConns(maxOpenConns)
	gdb.DB().SetMaxIdleConns(maxIdleConns)
	gdb.DB().SetConnMaxLifetime(time.Second * time.Duration(maxWaitTimeout))
	Db = gdb
	return err
}
