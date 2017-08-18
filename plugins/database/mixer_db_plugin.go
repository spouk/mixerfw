//---------------------------------------------------------------------------
//  Плагин для работы с базой данных для Mixerfw в виде миддла
//  Автор: spouk [spouk@spouk.ru] https://www.spouk.ru
//---------------------------------------------------------------------------
package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
	//_ "github.com/jinzhu/gorm/dialects/mssql"
	//_ "github.com/jinzhu/gorm/dialects/postgres"
	//_ "github.com/mxk/go-sqlite/sqlite3"
	//_ "github.com/lib/pq"
	//_ "github.com/mattn/go-sqlite3"
	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/denisenkom/go-mssqldb"
	"fmt"
	"mixerfw/fwpack"
	"time"
	"crypto/md5"
	"log"
)
type (
	//---------------------------------------------------------------------------
	//  DBS: основная структура
	//---------------------------------------------------------------------------
	MixerDBS struct {
		DB *gorm.DB
		dbinfo DBSInfoDatabase
		//logger Mixer.MixerLogger
		logger *log.Logger
	}
	//---------------------------------------------------------------------------
	//  DBSInfoDatabase: информационная структура для DBS
	//---------------------------------------------------------------------------
	DBSInfoDatabase struct {
		TypeDB string // `mysql`,`sqlite3`,`mssql`,`postgres`
		Host string
		Port int
		User string
		Password string
		Database string
		SSLMode bool
		SetMaxIdleConns int
		SetMaxOpenConns int
		SetConnMaxLifetime int
	}
)
//---------------------------------------------------------------------------
//  DBS: функционал
//---------------------------------------------------------------------------
func NewDBS(dbinfo DBSInfoDatabase, logger *log.Logger) *MixerDBS {
	dbs := new(MixerDBS)
	dbs.logger = logger
	db, err := gorm.Open(dbinfo.TypeDB, dbs.dns(dbinfo))
	fmt.Printf("DBS: %#v : %#v\n", db, err)
	if err != nil {
		if dbs.logger != nil {
			dbs.logger.Fatal(err.Error())
		} else {
			fmt.Printf("[dbs][error] %v\n", err)
		}
		return nil
	} else {
		dbs.DB = db
		db.DB().SetConnMaxLifetime(time.Duration(dbinfo.SetConnMaxLifetime) * time.Minute)
		db.DB().SetMaxIdleConns(dbinfo.SetMaxIdleConns)
		db.DB().SetMaxOpenConns(dbinfo.SetMaxOpenConns)
		return dbs
	}
}
func (d *MixerDBS) dns(dbinfo DBSInfoDatabase) string {
	var result string
	switch dbinfo.TypeDB {
	case "mysql":
		result = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local",
			dbinfo.User, dbinfo.Password, dbinfo.Database)
	case "sqlite3":
		result = dbinfo.Database
	case "mssql":
		result = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		dbinfo.User, dbinfo.Password,dbinfo.Host, dbinfo.Port, dbinfo.Database)
	case "postgres":
		result = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s",
		dbinfo.Host, dbinfo.User, dbinfo.Database, dbinfo.SSLMode, dbinfo.Password)
	}
	return result
}
func (d *MixerDBS) CreateTablesInDatabase(l []interface{}) {
	for _, x:=range l {
		d.logger.Printf("table: %v:%v\n", x, d.DB.HasTable(x))
		if !d.DB.HasTable(x) {
			d.DB.CreateTable(x)
		}
	}
}
func (d *MixerDBS) Cookgeneratenew(secret string) string {
	//генерация нового хэша
	t := time.Now()
	result := fmt.Sprintf("%x", md5.Sum([]byte(t.String() + secret)))
	return result
}
//---------------------------------------------------------------------------
//  DBS: middleware for mixer framework
//---------------------------------------------------------------------------
func (d *MixerDBS) DBSMixerMiddleware(handler Mixer.MixerHandler) Mixer.MixerHandler {
	return Mixer.MixerHandler(func(c *Mixer.MixerCarry) error {
		c.Set("db", d)
		handler(c)
		return nil
	})
}
