/*
Copyright 2024 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
)

type DBTYPE string
type CACHETYPE string

const (
	MYSQL    DBTYPE = "mysql"
	SQLITE3  DBTYPE = "sqlite3"
	POSTGRES DBTYPE = "postgres"

	CACHEREDIS  CACHETYPE = "redis"
	CACHEMEMORY CACHETYPE = "memory"
)

type CacheModel string

const (
	ORMCacheDisable     CacheModel = "CacheDisable"
	ORMCacheOnlyPrimary CacheModel = "CacheOnlyPrimary"
	ORMCacheOnlySearch  CacheModel = "CacheOnlySearch"
	ORMCacheAll         CacheModel = "CacheAll"
)

func (c CacheModel) Number() int {
	switch c {
	case ORMCacheDisable:
		return 0
	case ORMCacheOnlyPrimary:
		return 1
	case ORMCacheOnlySearch:
		return 2
	case ORMCacheAll:
		return 3
	default:
		return 3
	}
}

type OrmCacheCfg struct {
	// host:port address.
	Addr string `toml:"orm_cache_addr" json:"orm_cache_addr" env:"ORM_CACHE_ADDR"`
	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string `toml:"orm_cache_username" json:"orm_cache_username" env:"ORM_CACHE_USERNAME"`
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string `toml:"orm_cache_password" json:"orm_cache_password" env:"ORM_CACHE_PASSWORD"`
	// Database to be selected after connecting to the server.
	DB int `toml:"orm_cache_dbid" json:"orm_cache_dbid" env:"ORM_CACHE_DBID"`
}

type OrmCache struct {
	// cache type: redis, memory
	CacheType CACHETYPE `toml:"orm_cache_type" json:"orm_cache_type" env:"ORM_CACHE_TYPE" envDefault:"memory"`

	// 4 kinds of cache model
	CacheModel CacheModel `toml:"orm_cache_model" json:"orm_cache_model" env:"ORM_CACHE_MODEL" envDefault:"CacheOnlySearch"`

	// if user update/delete/create something in DB, we invalidate all cached data to ensure consistency,
	// else we do nothing to outdated cache.
	InvalidateWhenUpdate bool `toml:"orm_invalidate_when_update" json:"orm_invalidate_when_update" env:"ORM_INVALIDATE_WHEN_UPDATE" envDefault:"true"`

	// AsyncWrite if true, then we will write cache in async mode
	AsyncWrite bool `toml:"orm_async_write" json:"orm_async_write" env:"ORM_ASYNC_WRITE" envDefault:"true"`

	// CacheTTL cache ttl in ms, where 0 represents forever
	CacheTTL int64 `toml:"orm_cache_ttl" json:"orm_cache_ttl" env:"ORM_CACHE_TTL" envDefault:"5000"`

	// DisableCachePenetration if true, then we will not cache nil result
	DisableCachePenetrationProtect bool `toml:"orm_disable_cache_penetration_protect" json:"orm_disable_cache_penetration_protect" env:"ORM_DISABLE_CACHE_PENETRATION_PROTECT" envDefault:"false"`

	CacheCfg OrmCacheCfg `toml:"dbcachecfg"`
}

// DBConfig represents a database configuration
type DBConfig struct {
	DBType   DBTYPE   `toml:"db_type" json:"db_type" env:"ORM_DBTYPE"`     // dbtype
	User     string   `toml:"user" json:"user" env:"ORM_USER"`             // user
	Password string   `toml:"password" json:"password" env:"ORM_PASSWORD"` // password
	Host     string   `toml:"host" json:"host" env:"ORM_HOST"`             // host
	Port     string   `toml:"port" json:"port" env:"ORM_PORT"`             // port
	Database string   `toml:"database" json:"database" env:"ORM_DATABASE"` // database
	Args     string   `toml:"orm_args" json:"orm_args" env:"ORM_ARGS"`     // args
	Cache    OrmCache `toml:"dbcache"`
}

func (db DBConfig) TOML() string {
	return fmt.Sprintf(`
[database]
  ## dbtype 模式: mysql/sqlite3/postgres
  db_type = "%v"
  ## user 用户
  user = "%s"
  ## password 密码
  password = "%s"
  ## Host 地址
  host = "%s"
  ## Port 端口
  port = "%s"
  ## 数据库名
  database = "%s"
  ## args 参数
  args = "%s"
[dbcache]
  ## orm_cache_type: redis, memory. 默认值为memory
  orm_cache_type: %s
  ## orm_cache_model 模式: CacheDisable, CacheOnlyPrimary, CacheOnlySearch 和 CacheAll. 默认值为CacheOnlySearch
  orm_cache_model = %s
  ## orm_invalidate_when_update: 如果用户在数据库中更新/删除/创建某些内容，我们将使所有缓存数据无效以确保一致性. 如果true将会清理cache. 默认值为true
  orm_invalidate_when_update = %v
  ## orm_async_write: 如果 true，那么我们将以异步模式写入缓存. 默认值为true
  orm_async_write = %v
  ## orm_cache_ttl: CacheTTL 缓存 ttl，单位为 ms，其中 0 代表永远. 默认值为5000
  orm_cache_ttl = %v
  ## orm_disable_cache_penetration_protect: 如果为 true，那么我们不会缓存 nil 结果, 默认值为false
  orm_disable_cache_penetration_protect = %v
[dbcachecfg]
  ## cache addr
  orm_cache_addr = %v
  ## cache username
  orm_cache_username = %v
  ## cache password
  orm_cache_password = %v
  ## cache db id
  orm_cache_dbid = %v`,
		string(db.DBType),
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
		db.Args,
		string(db.Cache.CacheType),
		string(db.Cache.CacheModel),
		db.Cache.InvalidateWhenUpdate,
		db.Cache.AsyncWrite,
		db.Cache.CacheTTL,
		db.Cache.DisableCachePenetrationProtect,
		db.Cache.CacheCfg.Addr,
		db.Cache.CacheCfg.Username,
		db.Cache.CacheCfg.Password,
		db.Cache.CacheCfg.DB)
}

func (db DBConfig) DefaultConfig() DBConfig {
	db = DBConfig{
		DBType:   MYSQL,
		User:     "root",
		Password: "root",
		Host:     "localhost",
		Port:     "3306",
		Database: "test",
		Cache: OrmCache{
			CacheType:                      CACHETYPE("memory"),
			CacheModel:                     ORMCacheOnlySearch,
			InvalidateWhenUpdate:           true,
			CacheTTL:                       5000,
			AsyncWrite:                     true,
			DisableCachePenetrationProtect: false,
			CacheCfg: OrmCacheCfg{
				Addr:     "",
				Username: "",
				Password: "",
				DB:       0,
			},
		},
	}
	return db
}
