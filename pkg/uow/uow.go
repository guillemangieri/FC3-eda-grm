package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type RepositoryFactory func(tx *sql.Tx) interface{}

type UowInterface interface {
	Register(name string, fc RepositoryFactory)
	GetRepository(ctx context.Context, name string) (interface{}, error)
	Do(ctx context.Context, fn func(uow *Uow) error) error
	CommitOrRollback() error
	Rollback() error
	UnRegister(name string)
}

type Uow struct {
	Db           *sql.DB
	Tx           *sql.Tx
	Repositories map[string]RepositoryFactory
}

func NewUow(ctx context.Context, db *sql.DB) *Uow {
	return &Uow{
		Db:           db,
		Repositories: make(map[string]RepositoryFactory),
	}
}

func (u *Uow) Register(name string, fc RepositoryFactory) {
	u.Repositories[name] = fc
}

func (u *Uow) UnRegister(name string) {
	delete(u.Repositories, name)
}

func (u *Uow) GetRepository(ctx context.Context, name string) (interface{}, error) {
	if u.Tx == nil {
		return nil, errors.New("repository requested outside a transaction")
	}
	factory, ok := u.Repositories[name]
	if !ok {
		return nil, fmt.Errorf("repository %q is not registered", name)
	}
	repo := factory(u.Tx)
	return repo, nil
}

func (u *Uow) Do(ctx context.Context, fn func(Uow *Uow) error) error {
	tx, err := u.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txUow := &Uow{Db: u.Db, Tx: tx, Repositories: u.Repositories}
	err = fn(txUow)
	if err != nil {
		errRb := tx.Rollback()
		if errRb != nil {
			return errors.New(fmt.Sprintf("original error: %s, rollback error: %s", err.Error(), errRb.Error()))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		if errRb := tx.Rollback(); errRb != nil && !errors.Is(errRb, sql.ErrTxDone) {
			return fmt.Errorf("original error: %s, rollback error: %s", err, errRb)
		}
		return err
	}
	return nil
}

func (u *Uow) Rollback() error {
	if u.Tx == nil {
		return errors.New("no transaction to rollback")
	}
	err := u.Tx.Rollback()
	if err != nil {
		return err
	}
	u.Tx = nil
	return nil
}

func (u *Uow) CommitOrRollback() error {
	err := u.Tx.Commit()
	if err != nil {
		errRb := u.Rollback()
		if errRb != nil {
			return errors.New(fmt.Sprintf("original error: %s, rollback error: %s", err.Error(), errRb.Error()))
		}
		return err
	}
	u.Tx = nil
	return nil
}
