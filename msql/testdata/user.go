package testdata

import (
	"context"
	"time"
)

//Gender 性别
type Gender uint8

const (
	//GenderUnknown 性别,未知
	GenderUnknown Gender = iota
	//GenderFemale 性别,女
	GenderFemale
	//GenderMale 性别,男
	GenderMale
)

//User 用户信息
type User struct {
	Id        int64
	Name      string
	Gender    Gender
	Birthday  time.Time
	CreatedAt time.Time
}

//UserDao 用户信息Dao
//sql:table name=`user` dialect="mysql"
type UserDao interface {
	//FindById 通过ID获取用户信息
	/*sql:select query="select * from `user` where id = :id" master*/
	FindById(ctx context.Context, id int64) (*User, error)
	/*sql:select query="select * from `user` where birthday >= :time"*/
	FindByBirthdayGte(ctx context.Context /*sql:param ctx*/, time time.Time) ([]*User, error)
	/*sql:select query="select count(*) as count from `user` where birthday >= :time"*/
	CountByBirthdayGte(ctx context.Context /*sql:param ctx*/, time time.Time) (int32, error)
	//FindByName 通过用户名获取用户信息
	FindByName(ctx context.Context, name string) *User
	Insert(ctx context.Context, user *User) *User
	UpdateById(ctx context.Context, id int64, user *User) int64
	DeleteById(ctx context.Context, id int64) int64
}

//AToGender 字符转Gender
func AToGender(a string) Gender {
	switch a {
	case "FEMALE":
		//Don't parse comment
		return GenderFemale
	case "MALE":
		return GenderMale
	default:
		return GenderUnknown
	}
}
