package main

import (
	"log"
	"strconv"
	"time"

	"github.com/ogier/pflag"
)

func daily() {
	var day string
	pflag.StringVar(&day, "day",
		time.Now().AddDate(0, 0, -1).Format(DATEFORMAT),
		"compile stats from which day?")
	pflag.Parse()

	log.Print("-- running daily routine for ", day, ".")

	var sites []Site
	if err = pg.Model(Site{}).Scan(&sites); err != nil {
		log.Fatal("error fetching list of sites from postgres: ", err)
	}

	for _, site := range sites {
		log.Print("-------------")
		log.Print(" > site ", site.Code, " (", site.Name, "), from ", site.UserId, ":")
		key := redisKeyFactory(site.Code, day)

		stats := Compendium{
			Id:       makeBaseKey(site.Code, day),
			Sessions: make(map[string]string),
			Pages:    make(map[string]int),
		}

		// grab stats from redis
		if val, err := rds.HGetAll(key("s")).Result(); err == nil {
			for k, v := range val {
				stats.Sessions[k] = v
			}
		}
		if val, err := rds.HGetAll(key("p")).Result(); err == nil {
			for k, v := range val {
				if count, err := strconv.Atoi(v); err == nil {
					stats.Pages[k] = count
				}
			}
		}

		log.Print(stats)

		// check for zero-stats (to save disk space we won't store these)
		if len(stats.Sessions) == 0 && len(stats.Pages) == 0 {
			log.Print("   : skipped saving because everything is zero.")
			continue
		}

		// save on couch
		if _, err = couch.Put(stats.Id, stats, ""); err != nil {
			log.Print("   : failed to save stats on couch: ", err)
			continue
		}
		log.Print("   : saved on couch.")
	}
}

func monthly() {}
