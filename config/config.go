package config

type DbConfigStruct struct {
	User           string
	Password       string
	DBName         string
	Port           string
	MaxConnections int
}

// var DbConfig = DbConfigStruct{
// 	User:           "forum_user",
// 	Password:       "password",
// 	DBName:         "forum",
// 	Port:           "5432",
// 	MaxConnections: 1000,
// }

var DbConfig = DbConfigStruct{
	User:           "docker",
	Password:       "docker",
	DBName:         "docker",
	Port:           "5432",
	MaxConnections: 1000,
}
