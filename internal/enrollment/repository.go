package enrollment

import (
	"context"
	"log"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, e *domain.Enrollment) error
	GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Enrollment, error) //Le agregamos que getAll reciba filtros
	Count(ctx context.Context, filtros Filtros) (int, error)
	Update(ctx context.Context, id string, status *string) error
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

func (r *repo) GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Enrollment, error) {
	r.log.Println("repository GetAll:")

	var allEnroll []domain.Enrollment

	tx := r.db.WithContext(ctx).Model(&allEnroll)
	tx = aplicarFiltros(tx, filtros)
	tx = tx.Limit(limit).Offset(offset)
	result := tx.Order("created_at desc").Find(&allEnroll)
	if result.Error != nil {
		r.log.Println(result.Error)
		return nil, result.Error
	}
	r.log.Printf("all enrollments retrieved, rows affected: %d\n", result.RowsAffected)
	return allEnroll, nil
}

func aplicarFiltros(tx *gorm.DB, filtros Filtros) *gorm.DB {
	//Si el filtro es distinto de blanco (osea VIENE con filtro), le agregaremos un fultros
	if filtros.CourseID != "" {
		tx = tx.Where("course_id = ?", filtros.CourseID)
	}

	if filtros.UserID != "" {
		tx = tx.Where("user_id = ?", filtros.UserID)
	}
	return tx
}

func (r *repo) Count(ctx context.Context, filtros Filtros) (int, error) {
	var cantidad int64
	tx := r.db.WithContext(ctx).Model(domain.Enrollment{})
	tx = aplicarFiltros(tx, filtros)
	tx = tx.Count(&cantidad)
	if tx.Error != nil {
		r.log.Println(tx.Error)
		return 0, tx.Error
	}

	return int(cantidad), nil
}

func (r *repo) Update(ctx context.Context, id string, status *string) error {
	r.log.Println("repository Update")
	//Usaremos un MAP, porque si usamos el struct, NO ACTUALIZA VALORES CERO (osea "", 0, false)
	//Al usar un map es [string]intareface{}, se usa interface en el valor porque peude ser numerico, string, bool
	valores := make(map[string]interface{})

	if status != nil {
		valores["status"] = *status
	}

	result := r.db.WithContext(ctx).Model(domain.Enrollment{}).Where("id = ?", id).Updates(valores)

	if result.Error != nil {
		//Tambien imprimieros los errores en esta capa, ya no imprimiermos en la capa servicio
		r.log.Println(result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Printf("enrollment with id: %s not found, rows affected: %d\n", id, result.RowsAffected)
		return ErrEnrollNotFound{id}
	}
	r.log.Printf("enrollment updated with id: %s, rows affected: %d\n", id, result.RowsAffected)

	return nil
}
