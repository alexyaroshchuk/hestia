package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hestia/pkg/utils/parsers"
	"net/http"
	"sync"
	"time"

	"hestia/pkg/custerrors"
	"hestia/pkg/models"
	"hestia/pkg/repos"
)

var (
	ErrDuplicateFlat = errors.New("duplicate flat")
	FlatNotFound     = errors.New("flat not found")
)

type FlatInterface interface {
	Get(w http.ResponseWriter, r *http.Request) error
	GetAll(w http.ResponseWriter, r *http.Request) error
	Put(w http.ResponseWriter, r *http.Request) error
	Post(w http.ResponseWriter, r *http.Request) error
	Delete(w http.ResponseWriter, r *http.Request) error
}

// FlatService is the type that provides the main rules for flats.
type FlatService struct {
	rep        *repos.Store
	wg         *sync.WaitGroup
	collector  *parsers.Collector
	errHandler ErrFunc

	// NowFunc is used to get the current time.
	// Exposed for testing purposes.
	NowFunc func() time.Time
}

// NewFlatService creates a new Service.
func NewFlatService(db *sql.DB, collector *parsers.Collector, errHandler ErrFunc) *FlatService {
	svc := &FlatService{
		rep:        repos.New(db),
		wg:         &sync.WaitGroup{},
		errHandler: errHandler,
		collector:  collector,

		NowFunc: time.Now,
	}

	return svc
}

func (s *FlatService) Get(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	flat, err := s.rep.GetFlatByID(r.Context(), id)
	if err != nil {
		s.errHandler(err)
		return err
	}

	j, err := json.Marshal(flat)
	if err != nil {
		s.errHandler(err)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(j)
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *FlatService) GetAll(w http.ResponseWriter, r *http.Request) error {
	flats, err := s.rep.FindFlats(r.Context(), models.FlatFilter{})
	if err != nil {
		s.errHandler(err)
		return err
	}

	if len(flats) == 0 {
		s.errHandler(FlatNotFound)
		return custerrors.ErrNotFound
	}

	j, err := json.Marshal(flats)
	if err != nil {
		s.errHandler(err)
		return err
	}
	_, err = w.Write(j)
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *FlatService) Put(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	var flat models.Flat
	err := json.NewDecoder(r.Body).Decode(&flat)
	if err != nil {
		s.errHandler(err)
		return err
	}
	err = s.inTx(r.Context(), func(tx models.Tx) error {
		flat.ID = id
		err := tx.UpdateFlat(flat)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *FlatService) Post(w http.ResponseWriter, r *http.Request) error {
	var url models.Url
	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil {
		s.errHandler(err)
		return err
	}
	flat, err := s.collector.Parse(url.Url)
	if err != nil {
		s.errHandler(err)
		return err
	}

	err = s.inTx(r.Context(), func(tx models.Tx) error {
		err := tx.CreateFlat(flat)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *FlatService) Delete(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	err := s.inTx(r.Context(), func(tx models.Tx) error {
		err := tx.DeleteFlat(id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.errHandler(err)
		return err
	}

	return nil
}

func (s *FlatService) inTx(ctx context.Context, f func(tx models.Tx) error) error {
	tx, err := s.rep.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		rBackErr := tx.Rollback()
		if rBackErr != nil {
			err = errors.Join(err, rBackErr)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
