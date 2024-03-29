package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type File struct {
	ID         int
	Name       string
	Path       string
	UploadDate time.Time
}

var DB *sql.DB
var dbUser string = "root@tcp(localhost:3306)/"
var DbName string

func SetDB(httpPort string) {

	var err error
	DB, err = sql.Open("mysql", dbUser)
	if err != nil {
		fmt.Printf("mysql 접속 실패:%s\n", err)
		os.Exit(1)
	}

	DbName = "files_" + httpPort
	_, err = DB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", DbName))
	if err != nil {
		fmt.Printf("데이터베이스 생성 실패:%s\n", err)
		os.Exit(1)
	}

	DB, err = sql.Open("mysql", dbUser+DbName+"?parseTime=true")
	if err != nil {
		fmt.Printf("데이터베이스 연결 실패:%s\n", err)
		os.Exit(1)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS files (
		ID INTEGER PRIMARY KEY AUTO_INCREMENT,
		Name TEXT,
		Path TEXT,
		UploadDate TIMESTAMP
	)`)
	if err != nil {
		fmt.Printf("데이터베이스 테이블 생성 실패:%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("데이터베이스 연결 완료\n")
}

func GetFiles() ([]File, error) {
	rows, err := DB.Query("SELECT * FROM files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		err := rows.Scan(&file.ID, &file.Name, &file.Path, &file.UploadDate)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

func GetFileByName(fileName string) (*File, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := "SELECT ID, Name, Path, UploadDate FROM files WHERE NAME = ?"
	row := DB.QueryRow(query, fileName)

	var file File
	err := row.Scan(&file.ID, &file.Name, &file.Path, &file.UploadDate)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &file, nil
}

func SaveFileToDB(name string, path string) error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	uploadDate := time.Now()

	query := "INSERT INTO files (Name, Path, UploadDate) VALUES (?, ?, ?)"
	_, err := DB.Exec(query, name, path, uploadDate)
	if err != nil {
		return fmt.Errorf("error inserting file into database: %v", err)
	}

	fmt.Printf("File '%s' saved to the database.\n", name)
	return nil
}
