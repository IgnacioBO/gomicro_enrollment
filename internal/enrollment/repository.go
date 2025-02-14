package enrollment

import (
	"context"
	"log"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, e *domain.Enrollment) error
}

type repo struct {
	log *log.Logger
	db  *gorm.DB
}

func NewRepo(log *log.Logger, db *gorm.DB) Repository {
	return &repo{
		log: log,
		db:  db, //Devuevle un struct repo con la bbdd
	}

}

func (r *repo) Create(ctx context.Context, enrollment *domain.Enrollment) error {
	r.log.Println("repository Create:", enrollment)

	result := r.db.WithContext(ctx).Create(enrollment)

	if result.Error != nil {
		r.log.Println(result.Error)
		return result.Error
	}
	r.log.Printf("enrollment created with id: %s, rows affected: %d\n", enrollment.ID, result.RowsAffected)
	return nil
}
