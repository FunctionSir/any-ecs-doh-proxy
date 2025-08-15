/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2025-04-15 21:21:17
 * @LastEditTime: 2025-08-16 02:40:25
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /any-ecs-doh-proxy/dbops.go
 */
package main

import (
	"database/sql"
	"math/rand/v2"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Range struct {
	Begin int64
	End   int64
}

var DB *sql.DB
var GET_IP_STMT *sql.Stmt

var CachedIpRange map[string]map[string]map[string][]Range

func DbOpen() {
	var err error
	DB, err = sql.Open("sqlite3", Config["options"]["IpDb"])
	if err != nil {
		panic(err)
	}
}

func DbPrepare() {
	var err error
	GET_IP_STMT, err = DB.Prepare("SELECT BEGIN,END FROM IPV4 WHERE CODE=? AND PROVINCE=? AND CITY=?")
	if err != nil {
		panic(err)
	}
}

func getReadyForPos(countryCode, province, city string) {
	NewPosLock.Lock()
	defer NewPosLock.Unlock()
	if PosSet.Has(countryCode + "/" + province + "/" + city) {
		return
	}
	if CachedIpRange == nil {
		CachedIpRange = make(map[string]map[string]map[string][]Range)
		DnsCache = make(map[string]map[string]map[string]*sync.Map)
	}
	if CachedIpRange[countryCode] == nil {
		CachedIpRange[countryCode] = make(map[string]map[string][]Range)
		DnsCache[countryCode] = make(map[string]map[string]*sync.Map)
	}
	if CachedIpRange[countryCode][province] == nil {
		CachedIpRange[countryCode][province] = make(map[string][]Range)
		DnsCache[countryCode][province] = make(map[string]*sync.Map)
	}
	if CachedIpRange[countryCode][province][city] == nil {
		CachedIpRange[countryCode][province][city] = make([]Range, 0)
		DnsCache[countryCode][province][city] = &sync.Map{}
	}
	PosSet.Insert(countryCode + "/" + province + "/" + city)
}

func getIp(countryCode, province, city string) []byte {
	x := CachedIpRange[countryCode][province][city]
	if len(x) == 0 {
		result, err := GET_IP_STMT.Query(countryCode, province, city)
		if err != nil {
			return nil
		}
		for result.Next() {
			var err error
			tmpBegin := ""
			tmpEnd := ""
			result.Scan(&tmpBegin, &tmpEnd)
			tmp := Range{}
			tmp.Begin, err = strconv.ParseInt(tmpBegin, 10, 64)
			if err != nil {
				continue
			}
			tmp.End, err = strconv.ParseInt(tmpEnd, 10, 64)
			if err != nil {
				continue
			}
			x = append(x, tmp)
		}
	}
	if len(x) == 0 {
		return nil
	}
	choice := rand.IntN(len(x))
	gened := x[choice].End - 1 - rand.Int64N(x[choice].End-x[choice].Begin-2)
	if gened < 0 {
		return nil
	}
	b := [4]byte{byte(gened >> 24), byte(gened >> 16), byte(gened >> 8), byte(gened)}
	return b[:]
}
