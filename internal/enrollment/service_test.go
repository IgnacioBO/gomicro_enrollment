package enrollment_test

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	courseSdk "github.com/IgnacioBO/go_micro_sdk/course"
	userSdk "github.com/IgnacioBO/go_micro_sdk/user"

	courseSdkMock "github.com/IgnacioBO/go_micro_sdk/course/mock"
	userSdkMock "github.com/IgnacioBO/go_micro_sdk/user/mock"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"github.com/IgnacioBO/gomicro_enrollment/internal/enrollment"
)

func TestService_GetAll(t *testing.T) {

	//Aca generaremos un log par pasarselo / NO ES NECESARIO
	//io.Discard es para que no imprima nada en la consola, porque no queremos ver los logs en este test
	//entonecs para que creamos un log? Porque el servicio lo necesita, y si no se lo pasamos nos va a dar error SI PASSA POR UN print de log
	l := log.New(io.Discard, "", 0)

	//con t podemos encapsular los diferentes test que queramos hacer, hacer varios grupos o subconjuto de test usand el meotod Run
	//Aca hacemos un grupo
	t.Run("should return an error", func(t *testing.T) {
		//los want son los valores que esperamos que se devuelvan, en este caso un error
		//Generaremos una variable want que sera el error que esperamos que se devuelva
		var wantError error = errors.New("my error")

		//Tendremos tambie un conter y un wantCounter, estos sirve para testar que paso por aca (por ejemplo)
		var wantCounter int = 1
		var counter int = 0

		//Aca tenemos el repo del mock que creamos en el archivo mock_test.go
		//Aca implementamos el cuerpo de la funcion GetAllMock (porque en mock_test solo definimos la funcion GetAllMock, pero no la implementamos)
		//Y ponemos lo que queremos que devuelva cuando se llame a esa funcion, en este caso un error
		repo := &mockRepository{
			GetAllMock: func(ctx context.Context, filtros enrollment.Filtros, offset, limit int) ([]domain.Enrollment, error) {
				counter++
				return nil, errors.New("my error")
			},
		}

		//Entonces ahora instanciamos el servicio y le ponemos como parametro el repo que acabamos de crear
		svc := enrollment.NewService(l, nil, nil, repo)

		//Ahora con el svc ya creado, ejecutamos el getAll

		//context.Background() es un contexto vacio, porque no necesitamos pasarle nada, le pasamos tb filtros vacios y offset y limit en 0 y 10
		enrollments, err := svc.GetAll(context.Background(), enrollment.Filtros{}, 0, 10)
		//Y esperamos que el error sea ErrCourseIDRequired

		//Este assert verifica que sea un error
		assert.Error(t, err, "expected an error but got nil")

		//Aqui verifiicamos que el error sea el que esperamos, usando assert.EqualError
		//EqualErrors verifica que el mensaje del error sea igual al mensaje del error que esperamos
		assert.EqualError(t, err, wantError.Error(), "expected error to be '%s' but got '%s'", wantError.Error(), err.Error())

		//Este assert verifica que enrollments sea nil, porque no deberia devolver nada
		assert.Nil(t, enrollments, "expected enrollments to be nil but got a value")

		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
	})

	//Aca un test que deberia devolver un slice de enrollments
	t.Run("should return enrollments", func(t *testing.T) {
		//Aca generamos un slice de enrollments que queremos que se devuelva
		wantEnrollments := []domain.Enrollment{
			{ID: "1", UserID: "user1", CourseID: "course1", Status: domain.Pending},
			{ID: "2", UserID: "user2", CourseID: "course2", Status: domain.Active},
		}

		var wantCounter int = 1
		var counter int = 0

		repo := &mockRepository{
			GetAllMock: func(ctx context.Context, filtros enrollment.Filtros, offset, limit int) ([]domain.Enrollment, error) {
				counter++
				return []domain.Enrollment{
					{ID: "1", UserID: "user1", CourseID: "course1", Status: domain.Pending},
					{ID: "2", UserID: "user2", CourseID: "course2", Status: domain.Active},
				}, nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)

		enrollments, err := svc.GetAll(context.Background(), enrollment.Filtros{}, 0, 10)

		assert.NoError(t, err, "expected no error but got %v", err)
		assert.NotNil(t, enrollments, "expected enrollments to not be nil")
		assert.Equal(t, wantEnrollments, enrollments, "expected enrollments to be %v but got %v", wantEnrollments, enrollments)
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
	})

}

func TestService_Update(t *testing.T) {
	l := log.New(io.Discard, "", 0)
	t.Run("should return invalid status error", func(t *testing.T) {
		status := "R"
		var expectedError error = enrollment.ErrInvalidStatus{status}

		repo := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)

		err := svc.Update(context.Background(), "1", &status)
		assert.ErrorIs(t, err, expectedError, "expected error to be %v but got %v", expectedError, err)
	})

	t.Run("should return error from repo", func(t *testing.T) {
		status := "P"
		id := "1"
		var expectedError error = enrollment.ErrEnrollNotFound{id}

		repo := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return enrollment.ErrEnrollNotFound{id}
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)

		err := svc.Update(context.Background(), "1", &status)
		assert.ErrorIs(t, err, expectedError, "expected error to be %v but got %v", expectedError, err)
	})

	//Aca un test que no deberia devolver error, porque el mock devuelve nil
	t.Run("should return nil error", func(t *testing.T) {
		wantStatus := "P"
		wantId := "1"
		var wantCounter int = 1
		var counter int = 0

		repo := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				counter++
				//Validamos que los datos enviados por parametro sean los correctos
				assert.Equal(t, wantId, id, "expected id to be '%s' but got '%s'", wantId, id)
				assert.Equal(t, wantStatus, *status, "expected status to be '%s' but got '%s'", wantStatus, *status)
				return nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)

		status := "P"
		id := "1"
		err := svc.Update(context.Background(), id, &status)
		assert.NoError(t, err, "expected no error but got %v", err)
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
	})

}

func TestService_Count(t *testing.T) {
	l := log.New(io.Discard, "", 0)

	t.Run("should return an error", func(t *testing.T) {
		var wantError error = errors.New("my error")

		repo := &mockRepository{
			CountMock: func(ctx context.Context, filtros enrollment.Filtros) (int, error) {
				return 0, errors.New("my error")
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)
		count, err := svc.Count(context.Background(), enrollment.Filtros{})
		assert.Error(t, err, "expected an error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Zero(t, count, "expected count to be 0 but got %d", count)

	})

	t.Run("should return correct count number", func(t *testing.T) {
		var wantCount int = 50

		repo := &mockRepository{
			CountMock: func(ctx context.Context, filtros enrollment.Filtros) (int, error) {
				return 50, nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)
		count, err := svc.Count(context.Background(), enrollment.Filtros{})
		assert.NoError(t, err, "expected no error but got %v", err)
		assert.Equal(t, wantCount, count, "expected count to be %d but got %d", wantCount, count)
		assert.NotZero(t, count, "expected count to be greater than 0 but got %d", count)
	})

}

func TestService_Create(t *testing.T) {

	l := log.New(io.Discard, "", 0)
	/*
		if reps.StatusCode == 404 {
			return nil, ErrNotFound{fmt.Sprintf("%s", dataResponse.Message)}
		}*/

	t.Run("should return user not found error", func(t *testing.T) {
		//Aca podemos crear un erro de cero o en este caso usamos un error de un sdk que ya tenemos
		var wantError error = userSdk.ErrNotFound{Message: "User not found"}

		var wantCounter int = 1
		var counter int = 0

		userSdk := &UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				counter++
				return nil, userSdk.ErrNotFound{Message: "User not found"}
			},
		}

		//Realmente no es necesario el mock de repository, porque por flujo no se va a llamar al repo si el user no se encuentra
		repo := &mockRepository{
			CreateMock: func(ctx context.Context, e *domain.Enrollment) error {
				counter++
				return nil
			},
		}

		svc := enrollment.NewService(l, userSdk, nil, repo)

		userid := "1"
		courseid := "5"
		enrollments, err := svc.Create(context.Background(), userid, courseid)

		assert.Error(t, err, "expected an error but got nil")
		assert.ErrorIs(t, err, wantError, "expected error to be %v but got %v", wantError, err)
		assert.Nil(t, enrollments, "expected enrollments to be nil but got a value")
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
	})

	t.Run("should return course not found error", func(t *testing.T) {
		var wantError error = courseSdk.ErrNotFound{Message: "Course not found"}

		var wantCounter int = 2
		var counter int = 0
		userSdk := &UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				counter++
				return &domain.User{}, nil
			},
		}

		courseSdk := &CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				counter++
				return nil, courseSdk.ErrNotFound{Message: "Course not found"}
			},
		}

		svc := enrollment.NewService(l, userSdk, courseSdk, nil)

		userid := "1"
		courseid := "5"
		enrollments, err := svc.Create(context.Background(), userid, courseid)

		assert.Error(t, err, "expected an error but got nil")
		assert.ErrorIs(t, err, wantError, "expected error to be %v but got %v", wantError, err)
		assert.Nil(t, enrollments, "expected enrollments to be nil but got a value")
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
	})

	t.Run("should return repo error", func(t *testing.T) {
		var wantError error = errors.New("repo error")

		var wantCounter int = 3
		var counter int = 0

		userSdk := &UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				counter++
				return &domain.User{}, nil
			},
		}

		courseSdk := &CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				counter++
				return &domain.Course{}, nil
			},
		}

		repo := &mockRepository{
			CreateMock: func(ctx context.Context, e *domain.Enrollment) error {
				counter++
				return errors.New("repo error")
			},
		}

		svc := enrollment.NewService(l, userSdk, courseSdk, repo)

		userid := "1"
		courseid := "5"
		enrollments, err := svc.Create(context.Background(), userid, courseid)

		assert.Error(t, err, "expected an error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Nil(t, enrollments, "expected enrollments to be nil but got a value")
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
	})

	t.Run("should create new enrollment", func(t *testing.T) {

		wantId := "123"
		wantCourseId := "course1"
		wantUserId := "user1"
		wantStatus := domain.Pending

		var wantEnrollments = &domain.Enrollment{
			ID:       wantId,
			UserID:   wantUserId,
			CourseID: wantCourseId,
			Status:   wantStatus,
		}

		var wantCounter int = 3
		var counter int = 0

		//Aqui para variar usaremos el mock desde userSdk y courseSdk en vez de lso creados en este proyecto
		userSdk := &userSdkMock.UserSdkMock{
			GetMock: func(id string) (*domain.User, error) {
				assert.Equal(t, wantUserId, id, "expected user ID to be '%s' but got '%s'", wantUserId, id)
				counter++
				return &domain.User{}, nil
			},
		}

		courseSdk := &courseSdkMock.CourseSdkMock{
			GetMock: func(id string) (*domain.Course, error) {
				assert.Equal(t, wantCourseId, id, "expected course ID to be '%s' but got '%s'", wantCourseId, id)
				counter++
				return &domain.Course{}, nil
			},
		}

		repo := &mockRepository{
			CreateMock: func(ctx context.Context, e *domain.Enrollment) error {
				e.ID = "123" // Simulamos que el repo asigna un ID al enrollment
				counter++
				return nil
			},
		}

		svc := enrollment.NewService(l, userSdk, courseSdk, repo)

		userid := "user1"
		courseid := "course1"
		enrollments, err := svc.Create(context.Background(), userid, courseid)

		assert.NoError(t, err, "expected no error but got %v", err)
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
		assert.NotNil(t, enrollments, "expected enrollments to not be nil")
		assert.Equal(t, wantEnrollments.ID, enrollments.ID, "expected enrollment ID to be '%s' but got '%s'", wantEnrollments.ID, enrollments.ID)
		assert.Equal(t, wantEnrollments.UserID, enrollments.UserID, "expected enrollment UserID to be '%s' but got '%s'", wantEnrollments.UserID, enrollments.UserID)
		assert.Equal(t, wantEnrollments.CourseID, enrollments.CourseID, "expected enrollment CourseID to be '%s' but got '%s'", wantEnrollments.CourseID, enrollments.CourseID)
		assert.Equal(t, wantEnrollments.Status, enrollments.Status, "expected enrollment Status to be '%s' but got '%s'", wantEnrollments.Status, enrollments.Status)
		assert.Equal(t, wantEnrollments, enrollments, "expected enrollment to be %v but got %v", wantEnrollments, enrollments)
	})
}
