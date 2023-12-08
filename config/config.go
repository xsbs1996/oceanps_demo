package config

import "github.com/xsbs1996/oceanps/oceanpsfuncs"

var OceanpsRedisConf = &oceanpsfuncs.RedisPushPull{
	Ip:       "127.0.0.1",
	Port:     "6379",
	DB:       0,
	Password: "",
}
