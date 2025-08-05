package enrollment_test

import (
	"context"
	"errors"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"github.com/IgnacioBO/gomicro_enrollment/internal/enrollment"
)

// Structura basica de un mock
// Esta struct va a tener diferentes funciones.
// En en reposiorio tenemos unterface que tiene los metodos Create, GetAll, Count, Update
// Entonces para que este struct pueda ser un mock del repositorio, tiene que tener esos mismos metodos
// Entonces vamos a crear funciones que representen esos metodos
type mockRepository struct {
	//Aca tendremos campos que representen los datos que queramos mockear, por ejemplo el getAll en una parte usa s.repo.GetAll le pasamos valores que queremos que devuelva y asi mockear
	//Para eso generaremos para cada ese metodo un campo de nuestra struct que sera funciones que devuelvan los mismo valores que cada metodo
	//Entonces
	CreateMock func(ctx context.Context, e *domain.Enrollment) error
	GetAllMock func(ctx context.Context, filtros enrollment.Filtros, offset, limit int) ([]domain.Enrollment, error)
	CountMock  func(ctx context.Context, filtros enrollment.Filtros) (int, error)
	UpdateMock func(ctx context.Context, id string, status *string) error
	//Y ahora en cada funcion de abajo retornamos la funcion que corresponde a ese metodo
}

// Todas estas funciones son las mismas que las del repositorio
// Entonces cuando se llame a estas funciones, se llamara a las funciones que definimos arriba
func (m *mockRepository) Create(ctx context.Context, e *domain.Enrollment) error {
	//Este if obliga a que se setee el campo CreateMock, si no se setea, devuelve un error, que es una buena practica para evitar errores de que no se setee el mock
	if m.CreateMock == nil {
		return errors.New("CreateMock is not set")
	}
	return m.CreateMock(ctx, e)
}

func (m *mockRepository) GetAll(ctx context.Context, filtros enrollment.Filtros, offset, limit int) ([]domain.Enrollment, error) {
	return m.GetAllMock(ctx, filtros, offset, limit)
}

func (m *mockRepository) Count(ctx context.Context, filtros enrollment.Filtros) (int, error) {
	return m.CountMock(ctx, filtros)
}

func (m *mockRepository) Update(ctx context.Context, id string, status *string) error {
	return m.UpdateMock(ctx, id, status)
}

//MOcks de sdk -> Lo implementamos en el SDK pero lo dejamos aqui para que sea mas facil de entender
//Primero vamos al sdk y vemos que tiene este interface:
/*
	Transport interface {
		Get(id string) (*domain.User, error)
	}
*/
//Asi que usand este interface, creamos un mock de ese interface
type UserSdkMock struct {
	GetMock func(id string) (*domain.User, error)
}

func (m *UserSdkMock) Get(id string) (*domain.User, error) {
	return m.GetMock(id)
}

//Lo mismo para el courseSdk
/*
	Transport interface {
		Get(id string) (*domain.Course, error)
	}
*/
type CourseSdkMock struct {
	GetMock func(id string) (*domain.Course, error)
}

func (m *CourseSdkMock) Get(id string) (*domain.Course, error) {
	return m.GetMock(id)
}
