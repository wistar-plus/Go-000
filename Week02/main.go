package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

/*
	问题：我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？
	答：应该wrap err并抛给上层，应为dao层已经属于业务层，业务层报错都应该wrap error抛给上层。如果不知到该不该报err可以提供一个方法给上层判断
*/

var testerr = sql.ErrNoRows

func main() {
	addr := "localhost:12340"

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if _, err := service(); err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
		fmt.Println("success")
	})

	go http.ListenAndServe(addr, nil)
	http.Get("http://" + addr + "/ping")
	testerr = sql.ErrConnDone
	http.Get("http://" + addr + "/ping")
	time.Sleep(time.Microsecond * 50)
}

func service() ([]int, error) {
	res, err := dao()
	if err != nil && IsNoRows(err) {
		res = []int{}
	} else if err != nil {
		return nil, err
	}
	return res, nil
}

func dao() ([]int, error) {
	err := testerr
	if err != nil {
		return nil, errors.Wrap(err, "sql query err")
	}
	return []int{}, nil
}

func IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
