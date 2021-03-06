// Code generated by meta sql:table,sql:select,sql:insert,sql:update,sql:delete,sql:none. DO NOT EDIT.

//go:build !ignore_meta_generated
// +build !ignore_meta_generated

package testdata

import (
	"context"
	"time"

	"github.com/gomelon/melon/data"
)

type UserDaoImpl struct {
	_tm *data.SQLTXManager
}

func NewUserDaoImpl(_tm *data.SQLTXManager) *UserDaoImpl {
	return &UserDaoImpl{
		_tm: _tm,
	}
}

func (_impl *UserDaoImpl) CountByBirthdayGTE(ctx context.Context, time time.Time) (int, error) {
	_sql := "SELECT COUNT(*) AS X FROM `user` WHERE (`birthday` >= ?)"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, time)

	var _item int
	if _err != nil {
		return _item, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _item, _rows.Err()
	}

	_item = int(0)
	_err = _rows.Scan(&_item)
	return _item, _err
}

func (_impl *UserDaoImpl) CountByBirthdayGTE2(ctx context.Context, time time.Time) (int, error) {
	_sql := "select count(*) as count from `user` where birthday >= ?"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, time)

	var _item int
	if _err != nil {
		return _item, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _item, _rows.Err()
	}

	_item = int(0)
	_err = _rows.Scan(&_item)
	return _item, _err
}

func (_impl *UserDaoImpl) DeleteById(ctx context.Context, id int64) (int64, error) {
	_sql := "DELETE FROM `user` WHERE (`id` = ?)"
	_result, err := _impl._tm.OriginTXOrDB(ctx).
		Exec(_sql, id)
	if err != nil {
		return 0, err
	}
	return _result.RowsAffected()
}

func (_impl *UserDaoImpl) DeleteById2(ctx context.Context, id int64) (int64, error) {
	_sql := "delete from `user` where id = ?"
	_result, err := _impl._tm.OriginTXOrDB(ctx).
		Exec(_sql, id)
	if err != nil {
		return 0, err
	}
	return _result.RowsAffected()
}

func (_impl *UserDaoImpl) ExistsById(ctx context.Context, id int64) (bool, error) {
	_sql := "SELECT 1 AS X FROM `user` WHERE (`id` = ?) LIMIT 0, 1"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, id)

	var _item bool
	if _err != nil {
		return _item, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _item, _rows.Err()
	}

	_item = false
	_err = _rows.Scan(&_item)
	return _item, _err
}

func (_impl *UserDaoImpl) ExistsById2(ctx context.Context, id int64) (bool, error) {
	_sql := "select 1 as X from `user` WHERE id = ? limit 1"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, id)

	var _item bool
	if _err != nil {
		return _item, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _item, _rows.Err()
	}

	_item = false
	_err = _rows.Scan(&_item)
	return _item, _err
}

func (_impl *UserDaoImpl) FindByBirthdayGTE(ctx context.Context, time time.Time) ([]*User, error) {
	_sql := "SELECT id, name, gender, birthday, created_at FROM `user` WHERE (`birthday` >= ?)"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, time)

	var _items []*User
	if _err != nil {
		return _items, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _items, _rows.Err()
	}

	for _rows.Next() {
		_item := &User{}
		_err = _rows.Scan(&_item.Id, &_item.Name, &_item.Gender, &_item.Birthday, &_item.CreatedAt)
		if _err != nil {
			return _items, _err
		}
		_items = append(_items, _item)
	}
	return _items, nil
}

func (_impl *UserDaoImpl) FindByBirthdayGTE2(ctx context.Context, time time.Time) ([]*User, error) {
	_sql := "select id, name, gender, birthday, created_at from `user` where birthday >= ?"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, time)

	var _items []*User
	if _err != nil {
		return _items, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _items, _rows.Err()
	}

	for _rows.Next() {
		_item := &User{}
		_err = _rows.Scan(&_item.Id, &_item.Name, &_item.Gender, &_item.Birthday, &_item.CreatedAt)
		if _err != nil {
			return _items, _err
		}
		_items = append(_items, _item)
	}
	return _items, nil
}

func (_impl *UserDaoImpl) FindById(ctx context.Context, id int64) (*User, error) {
	_sql := "SELECT id, name, gender, birthday, created_at FROM `user` WHERE (`id` = ?)"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, id)

	var _item *User
	if _err != nil {
		return _item, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _item, _rows.Err()
	}

	_item = &User{}
	_err = _rows.Scan(&_item.Id, &_item.Name, &_item.Gender, &_item.Birthday, &_item.CreatedAt)
	return _item, _err
}

func (_impl *UserDaoImpl) FindById2(ctx context.Context, id int64) (*User, error) {
	_sql := "select id, name, gender, birthday, created_at from `user` where id = ?"
	_rows, _err := _impl._tm.OriginTXOrDB(ctx).
		Query(_sql, id)

	var _item *User
	if _err != nil {
		return _item, _err
	}

	defer _rows.Close()

	if !_rows.Next() {
		return _item, _rows.Err()
	}

	_item = &User{}
	_err = _rows.Scan(&_item.Id, &_item.Name, &_item.Gender, &_item.Birthday, &_item.CreatedAt)
	return _item, _err
}
