package enrollment

import (
	"context"
	"log"

	"github.com/IgnacioBO/gomicro_domain/domain"

	courseSdk "github.com/IgnacioBO/go_micro_sdk/course"
	userSdk "github.com/IgnacioBO/go_micro_sdk/user"
)

type Service interface {
	Create(ctx context.Context, userID, courseID string) (*domain.Enrollment, error)
	GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Enrollment, error) //Le agregamos que getAll reciba filtros
	Count(ctx context.Context, Filtros Filtros) (int, error)
	Update(ctx context.Context, id string, status *string) error
}

type (
	Filtros struct {
		UserID   string
		CourseID string
	}
)

type service struct {
	log         *log.Logger
	userTrans   userSdk.Transport
	courseTrans courseSdk.Transport
	repo        Repository
}

func NewService(log *log.Logger, userTrans userSdk.Transport, courseTrans courseSdk.Transport, repo Repository) Service {
	return &service{
		log:         log,
		userTrans:   userTrans,
		courseTrans: courseTrans,
		repo:        repo,
	}
}

func (s service) Create(ctx context.Context, userID, courseID string) (*domain.Enrollment, error) {
	s.log.Println("Create enrollment service")

	enrollmentNuevo := &domain.Enrollment{
		UserID:   userID,
		CourseID: courseID,
		Status:   domain.Pending,
	}

	//Haremos los get de user y course, si da error devolvemos el error
	_, err := s.userTrans.Get(userID)
	if err != nil {
		return nil, err
	}

	if _, err := s.courseTrans.Get(courseID); err != nil {
		return nil, err
	}
	//Le pasamo al repo el domain.Course (del domain.go) a la capa repo a la funcion Create (que recibe puntero)
	err = s.repo.Create(ctx, enrollmentNuevo)
	//Si hay un error (por ejemplo al insertar, se devuelve el error y la capa endpoitn lo maneja con un status code y todo)
	if err != nil {
		return nil, err
	}
	return enrollmentNuevo, nil
}

func (s service) GetAll(ctx context.Context, filtros Filtros, offset, limit int) ([]domain.Enrollment, error) {
	s.log.Println("GetAll enrollment service")

	allEnroll, err := s.repo.GetAll(ctx, filtros, offset, limit)
	if err != nil {
		return nil, err
	}
	return allEnroll, nil
}

func (s service) Count(ctx context.Context, filtros Filtros) (int, error) {
	s.log.Println("Count users service")
	return s.repo.Count(ctx, filtros)
}

func (s service) Update(ctx context.Context, id string, status *string) error {
	s.log.Println("Update user service")

	if status != nil { //validamos que no venga nill (pq viene como puntero 'status *string)
		switch domain.EnrollStatus(*status) { //Aqui transforamos el status en domain.EnrollStatus
		case domain.Pending, domain.Active, domain.Studying, domain.Inactive:
		default:
			return ErrInvalidStatus{*status}
		}
	}

	err := s.repo.Update(ctx, id, status)
	return err
}
