package enrollment_test

//Test integrales -> Probaremos endpoints y servicios en conjuntos
//Pero mockearemos el repositorio y el sdk

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/IgnacioBO/go_lib_response/response"
	courseSdk "github.com/IgnacioBO/go_micro_sdk/course"
	userSdk "github.com/IgnacioBO/go_micro_sdk/user"

	courseSdkMock "github.com/IgnacioBO/go_micro_sdk/course/mock"
	userSdkMock "github.com/IgnacioBO/go_micro_sdk/user/mock"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"github.com/IgnacioBO/gomicro_enrollment/internal/enrollment"
)

func TestEndpoint_Create(t *testing.T) {

	l := log.New(io.Discard, "", 0)
	/*
		if reps.StatusCode == 404 {
			return nil, ErrNotFound{fmt.Sprintf("%s", dataResponse.Message)}
		}*/

	t.Run("should return user id required error", func(t *testing.T) {
		enrollmentEndpoint := enrollment.MakeEndpoints(nil, enrollment.Config{LimitPageDefault: "10"})
		userid := ""
		courseid := "course1"
		enrollmentsResponse, err := enrollmentEndpoint.Create(context.Background(), enrollment.CreateRequest{
			UserID:   userid,
			CourseID: courseid,
		})

		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, enrollment.ErrUserIDRequired.Error(), "expected error message to be '%s' but got '%s'", enrollment.ErrUserIDRequired.Error(), err.Error())
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusBadRequest, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return course id required error", func(t *testing.T) {
		enrollmentEndpoint := enrollment.MakeEndpoints(nil, enrollment.Config{LimitPageDefault: "10"})
		userid := "user1"
		courseid := ""
		enrollmentsResponse, err := enrollmentEndpoint.Create(context.Background(), enrollment.CreateRequest{
			UserID:   userid,
			CourseID: courseid,
		})

		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, enrollment.ErrCourseIDRequired.Error(), "expected error message to be '%s' but got '%s'", enrollment.ErrCourseIDRequired.Error(), err.Error())
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusBadRequest, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	//Test happy path y otros errores usando table driven tests o matrix tests
	//Generaremos una variable que tendra un slice de struct que definira los casos de prueba
	//Osea aqui podemos ir variando los mocks que simularemos en sdk y otros

	//Aqui repetimos varios test de integracion que tambien los hicimos dentro de unit test de service_test.go asi que no es necesario, pero lo dejamos para ver como se haria y repasar
	opciones := []struct {
		tag             string
		repositoryMock  enrollment.Repository
		userSdkMock     userSdk.Transport
		courseSdkMock   courseSdk.Transport
		wantError       error
		wantCode        int
		wantEnrollments *domain.Enrollment
	}{
		{
			tag: "should return user not found error",
			userSdkMock: &userSdkMock.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, userSdk.ErrNotFound{Message: "User not found"}
				},
			},
			wantError: userSdk.ErrNotFound{Message: "User not found"},
			wantCode:  http.StatusNotFound,
		},
		{
			tag: "should return course not found error",
			//Le pasamos user para que no de error al hacer el get del user
			userSdkMock: &userSdkMock.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return &domain.User{}, nil
				},
			},
			courseSdkMock: &courseSdkMock.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return nil, courseSdk.ErrNotFound{Message: "Course not found"}
				},
			},
			wantError: courseSdk.ErrNotFound{Message: "Course not found"},
			wantCode:  http.StatusNotFound,
		},
		{
			tag: "should return error different from not found",
			//Le pasamos user para que no de error al hacer el get del user
			userSdkMock: &userSdkMock.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, errors.New("some other error")
				},
			},
			wantError: errors.New("some other error"),
			wantCode:  http.StatusInternalServerError,
		},
		{
			tag: "should return error from repository",
			//Le pasamos user para que no de error al hacer el get del user
			userSdkMock: &userSdkMock.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, nil
				},
			},
			courseSdkMock: &courseSdkMock.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return &domain.Course{}, nil
				},
			},
			repositoryMock: &mockRepository{
				CreateMock: func(ctx context.Context, e *domain.Enrollment) error {
					return errors.New("error from repository")
				},
			},
			wantError: errors.New("error from repository"),
			wantCode:  http.StatusInternalServerError,
		},
		{
			tag: "should create new enrollment",
			userSdkMock: &userSdkMock.UserSdkMock{
				GetMock: func(id string) (*domain.User, error) {
					return nil, nil
				},
			},
			courseSdkMock: &courseSdkMock.CourseSdkMock{
				GetMock: func(id string) (*domain.Course, error) {
					return &domain.Course{}, nil
				},
			},
			repositoryMock: &mockRepository{
				CreateMock: func(ctx context.Context, e *domain.Enrollment) error {
					e.ID = "123" // Simulamos que el repo asigna un ID al enrollment
					return nil
				},
			},
			wantError: nil,
			wantCode:  http.StatusCreated,
			wantEnrollments: &domain.Enrollment{
				ID:       "123",
				UserID:   "user1",
				CourseID: "course1",
				Status:   domain.Pending,
			},
		},
	}

	for _, opcion := range opciones {
		t.Run(opcion.tag, func(t *testing.T) {
			svc := enrollment.NewService(l, opcion.userSdkMock, opcion.courseSdkMock, opcion.repositoryMock)
			enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{})
			//Ya que no probaremos distinso usuarios y cursos, podemos usar siempre los mismos (los probamos antes de esta matriz)
			userid := "user1"
			courseid := "course1"
			enrollmentsResponse, err := enrollmentEndpoint.Create(context.Background(), enrollment.CreateRequest{
				UserID:   userid,
				CourseID: courseid,
			})

			//Si esperamos un error haremos cosas y en el elsoe serian los casos que no esperamos error
			if opcion.wantError != nil {
				assert.Error(t, err, "expected error but got nil")
				assert.Nil(t, enrollmentsResponse, "expected enrollments to be nil")

				responseError := err.(response.Response)
				assert.Equal(t, opcion.wantCode, responseError.StatusCode(), "expected status code to be %d but got %d", opcion.wantCode, responseError.StatusCode())
				assert.EqualError(t, err, opcion.wantError.Error(), "expected error message to be '%s' but got '%s'", opcion.wantError.Error(), err.Error())

			} else {
				assert.NoError(t, err, "expected no error but got %v", err)
				assert.NotNil(t, enrollmentsResponse, "expected enrollments to not be nil")

				//El response tiene un campo Data que es el enrollment creado
				//Usamos .(*response.Response) para hacer un type assertion y obtener el Data (es como castear a un tipo concreto y poder acceder a sus campos)
				//puede ser -> enrollmentsResponse.(*response.SuccessResponse).Data.(*domain.Enrollment) o enrollmentsResponse.(response.Response).GetData().(*domain.Enrollment)
				enrollments := enrollmentsResponse.(response.Response).GetData().(*domain.Enrollment)
				statusCode := enrollmentsResponse.(response.Response).StatusCode()

				assert.NoError(t, err, "expected no error but got %v", err)
				assert.NotNil(t, enrollmentsResponse, "expected enrollments to not be nil")
				assert.Equal(t, opcion.wantEnrollments.ID, enrollments.ID, "expected enrollment ID to be '%s' but got '%s'", opcion.wantEnrollments.ID, enrollments.ID)
				assert.Equal(t, opcion.wantEnrollments.UserID, enrollments.UserID, "expected enrollment UserID to be '%s' but got '%s'", opcion.wantEnrollments.UserID, enrollments.UserID)
				assert.Equal(t, opcion.wantEnrollments.CourseID, enrollments.CourseID, "expected enrollment CourseID to be '%s' but got '%s'", opcion.wantEnrollments.CourseID, enrollments.CourseID)
				assert.Equal(t, opcion.wantEnrollments.Status, enrollments.Status, "expected enrollment Status to be '%s' but got '%s'", opcion.wantEnrollments.Status, enrollments.Status)
				assert.Equal(t, opcion.wantEnrollments, enrollments, "expected enrollment to be %v but got %v", opcion.wantEnrollments, enrollments)
				assert.Equal(t, opcion.wantCode, statusCode, "expected status code to be %d but got %d", http.StatusCreated, statusCode)
			}

		})
	}

	t.Run("should create new enrollment not matrix", func(t *testing.T) {

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
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})

		userid := "user1"
		courseid := "course1"
		// Esto es lo que se llama en el endpoint, por eso no pasamos el contexto, porque ya lo tenemos en la funcion
		//Devolvera un SuccessResponse con el enrollment creado
		enrollmentsResponse, err := enrollmentEndpoint.Create(context.Background(), enrollment.CreateRequest{
			UserID:   userid,
			CourseID: courseid,
		})

		//El succes response tiene un campo Data que es el enrollment creado
		//Usamos .(*response.SuccessResponse) para hacer un type assertion y obtener el Data (es como castear a un tipo concreto y poder acceder a sus campos)
		//puede ser -> enrollmentsResponse.(*response.SuccessResponse).Data.(*domain.Enrollment) o enrollmentsResponse.(response.Response).GetData().(*domain.Enrollment)
		enrollments := enrollmentsResponse.(response.Response).GetData().(*domain.Enrollment)
		//	println("Enrollments:", enrollments.(*domain.Enrollment).ID)
		println(wantEnrollments)

		assert.NoError(t, err, "expected no error but got %v", err)
		assert.Equal(t, wantCounter, counter, "expected counter to be %d but got %d", wantCounter, counter)
		assert.NotNil(t, enrollmentsResponse, "expected enrollments to not be nil")
		assert.Equal(t, wantEnrollments.ID, enrollments.ID, "expected enrollment ID to be '%s' but got '%s'", wantEnrollments.ID, enrollments.ID)
		assert.Equal(t, wantEnrollments.UserID, enrollments.UserID, "expected enrollment UserID to be '%s' but got '%s'", wantEnrollments.UserID, enrollments.UserID)
		assert.Equal(t, wantEnrollments.CourseID, enrollments.CourseID, "expected enrollment CourseID to be '%s' but got '%s'", wantEnrollments.CourseID, enrollments.CourseID)
		assert.Equal(t, wantEnrollments.Status, enrollments.Status, "expected enrollment Status to be '%s' but got '%s'", wantEnrollments.Status, enrollments.Status)
		assert.Equal(t, wantEnrollments, enrollments, "expected enrollment to be %v but got %v", wantEnrollments, enrollments)
	})

}

func TestEndpoint_GetAll(t *testing.T) {

	l := log.New(io.Discard, "", 0)
	/*
		if reps.StatusCode == 404 {
			return nil, ErrNotFound{fmt.Sprintf("%s", dataResponse.Message)}
		}*/

	t.Run("should return error because error on count", func(t *testing.T) {
		wantError := errors.New("error from repository")
		repositoryMock := &mockRepository{
			CountMock: func(ctx context.Context, filtros enrollment.Filtros) (int, error) {
				return 0, errors.New("error from repository")
			},
		}
		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{})
		enrollmentsResponse, err := enrollmentEndpoint.GetAll(context.Background(), enrollment.GetAllRequest{})

		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error message to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusInternalServerError, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return error because error on meta", func(t *testing.T) {
		wantError := errors.New("strconv.Atoi: parsing \"WrongPage\": invalid syntax")
		repositoryMock := &mockRepository{
			CountMock: func(ctx context.Context, filtros enrollment.Filtros) (int, error) {
				return 10, nil
			},
		}
		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "WrongPage"})
		enrollmentsResponse, err := enrollmentEndpoint.GetAll(context.Background(), enrollment.GetAllRequest{})

		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error message to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusInternalServerError, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return error because error on GetAll Repository", func(t *testing.T) {
		wantError := errors.New("repo error")

		repositoryMock := &mockRepository{
			CountMock: func(ctx context.Context, filtros enrollment.Filtros) (int, error) {
				return 10, nil
			},
			GetAllMock: func(ctx context.Context, filtros enrollment.Filtros, offset, limit int) ([]domain.Enrollment, error) {
				return nil, errors.New("repo error")
			},
		}

		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})
		userid := "user1"
		courseid := "course1"
		enrollmentsResponse, err := enrollmentEndpoint.GetAll(context.Background(), enrollment.GetAllRequest{
			UserID:   userid,
			CourseID: courseid,
			Limit:    20,
			Page:     1,
		})

		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error message to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusInternalServerError, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return enrollments", func(t *testing.T) {

		wantEnrollments := []domain.Enrollment{
			{ID: "1", UserID: "user1", CourseID: "course1", Status: domain.Pending},
			{ID: "2", UserID: "user2", CourseID: "course2", Status: domain.Active},
		}

		repo := &mockRepository{
			GetAllMock: func(ctx context.Context, filtros enrollment.Filtros, offset, limit int) ([]domain.Enrollment, error) {
				return []domain.Enrollment{
					{ID: "1", UserID: "user1", CourseID: "course1", Status: domain.Pending},
					{ID: "2", UserID: "user2", CourseID: "course2", Status: domain.Active},
				}, nil
			},
			CountMock: func(ctx context.Context, filtros enrollment.Filtros) (int, error) {
				return 2, nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repo)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})
		userid := "user1"
		courseid := "course1"
		enrollmentsResponse, err := enrollmentEndpoint.GetAll(context.Background(), enrollment.GetAllRequest{
			UserID:   userid,
			CourseID: courseid,
			Limit:    10,
			Page:     1,
		})

		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := enrollmentsResponse.(response.Response)
		enrollments := enrollmentsResponse.(response.Response).GetData().([]domain.Enrollment)

		assert.NoError(t, err, "expected no error but got %v", err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusOK, resp.StatusCode())
		assert.NotNil(t, enrollmentsResponse, "expected enrollments to not be nil")
		//Aqui no es necesario ir comparando uno a uno, si no el slice completo, pero por si a caso lo dejamos
		assert.Equal(t, wantEnrollments[0].ID, enrollments[0].ID, "expected first enrollment ID to be '%s' but got '%s'", wantEnrollments[0].ID, enrollments[0].ID)
		assert.Equal(t, wantEnrollments[0].UserID, enrollments[0].UserID, "expected first enrollment UserID to be '%s' but got '%s'", wantEnrollments[0].UserID, enrollments[0].UserID)
		assert.Equal(t, wantEnrollments[0].CourseID, enrollments[0].CourseID, "expected first enrollment CourseID to be '%s' but got '%s'", wantEnrollments[0].CourseID, enrollments[0].CourseID)
		assert.Equal(t, wantEnrollments[0].Status, enrollments[0].Status, "expected first enrollment Status to be '%s' but got '%s'", wantEnrollments[0].Status, enrollments[0].Status)
		assert.Equal(t, wantEnrollments[1].ID, enrollments[1].ID, "expected second enrollment ID to be '%s' but got '%s'", wantEnrollments[1].ID, enrollments[1].ID)
		assert.Equal(t, wantEnrollments[1].UserID, enrollments[1].UserID, "expected second enrollment UserID to be '%s' but got '%s'", wantEnrollments[1].UserID, enrollments[1].UserID)
		assert.Equal(t, wantEnrollments[1].CourseID, enrollments[1].CourseID, "expected second enrollment CourseID to be '%s' but got '%s'", wantEnrollments[1].CourseID, enrollments[1].CourseID)
		assert.Equal(t, wantEnrollments[1].Status, enrollments[1].Status, "expected second enrollment Status to be '%s' but got '%s'", wantEnrollments[1].Status, enrollments[1].Status)
		//Comparamos el slice completo
		assert.Equal(t, wantEnrollments, enrollments, "expected enrollments to be '%v' but got '%v'", wantEnrollments, enrollments)

	})

}

func TestEndpoint_Updated(t *testing.T) {

	l := log.New(io.Discard, "", 0)

	t.Run("should return error if status is empty", func(t *testing.T) {
		enrollmentEndpoint := enrollment.MakeEndpoints(nil, enrollment.Config{LimitPageDefault: "10"})
		id := "enrollment1"
		status := ""
		enrollmentsResponse, err := enrollmentEndpoint.Update(context.Background(), enrollment.UpdateRequest{
			ID:     id,
			Status: &status,
		})

		assert.Error(t, err, "expected error but got nil")
		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, enrollment.ErrStatusRequired.Error(), "expected error message to be '%s' but got '%s'", enrollment.ErrStatusRequired.Error(), err.Error())
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusBadRequest, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return error if status in longer than 2 characters", func(t *testing.T) {
		enrollmentEndpoint := enrollment.MakeEndpoints(nil, enrollment.Config{LimitPageDefault: "10"})
		id := "enrollment1"
		status := "longer than 2 characters"
		enrollmentsResponse, err := enrollmentEndpoint.Update(context.Background(), enrollment.UpdateRequest{
			ID:     id,
			Status: &status,
		})

		assert.Error(t, err, "expected error but got nil")
		//Para ovtener el statucs code, hay que castear el error a response.Response (porque asi lo definieos en el middlewaer de encodeError) y ahi usar el metodo StatusCode()
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, enrollment.ErrStatusTooLong.Error(), "expected error message to be '%s' but got '%s'", enrollment.ErrStatusTooLong.Error(), err.Error())
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusBadRequest, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return error if enrollment not found on repository", func(t *testing.T) {
		status := "P"
		id := "1"

		wantError := enrollment.ErrEnrollNotFound{id}

		repositoryMock := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return enrollment.ErrEnrollNotFound{id}
			},
		}

		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})
		enrollmentsResponse, err := enrollmentEndpoint.Update(context.Background(), enrollment.UpdateRequest{
			ID:     id,
			Status: &status,
		})

		assert.Error(t, err, "expected error but got nil")
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error message to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusNotFound, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return error if status in not valid on service", func(t *testing.T) {
		status := "J"
		id := "1"

		wantError := enrollment.ErrInvalidStatus{Status: status}

		repositoryMock := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})
		enrollmentsResponse, err := enrollmentEndpoint.Update(context.Background(), enrollment.UpdateRequest{
			ID:     id,
			Status: &status,
		})

		assert.Error(t, err, "expected error but got nil")
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error message to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusBadRequest, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should return error if service returns error from repository", func(t *testing.T) {
		status := "P"
		id := "1"

		wantError := errors.New("error from repo")

		repositoryMock := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				return errors.New("error from repo")
			},
		}

		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})
		enrollmentsResponse, err := enrollmentEndpoint.Update(context.Background(), enrollment.UpdateRequest{
			ID:     id,
			Status: &status,
		})

		assert.Error(t, err, "expected error but got nil")
		resp := err.(response.Response)

		assert.Error(t, err, "expected error but got nil")
		assert.EqualError(t, err, wantError.Error(), "expected error message to be '%s' but got '%s'", wantError.Error(), err.Error())
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusInternalServerError, resp.StatusCode())
		assert.Nil(t, enrollmentsResponse, "expected enrollments to not be nil")

	})

	t.Run("should update enrollment", func(t *testing.T) {
		status := "A"
		id := "1"

		repositoryMock := &mockRepository{
			UpdateMock: func(ctx context.Context, id string, status *string) error {
				assert.Equal(t, id, "1", "expected enrollment ID to be '%s' but got '%s'", id, "1")
				assert.Equal(t, *status, "A", "expected enrollment Status to be '%s' but got '%s'", *status, "A")
				assert.NotNil(t, status, "expected enrollment Status to not be nil")
				return nil
			},
		}

		svc := enrollment.NewService(l, nil, nil, repositoryMock)
		enrollmentEndpoint := enrollment.MakeEndpoints(svc, enrollment.Config{LimitPageDefault: "10"})
		enrollmentsResponse, err := enrollmentEndpoint.Update(context.Background(), enrollment.UpdateRequest{
			ID:     id,
			Status: &status,
		})

		assert.NotNil(t, enrollmentsResponse, "expected response to not be nil")
		resp := enrollmentsResponse.(response.Response)
		responseData := enrollmentsResponse.(response.Response).GetData().(enrollment.UpdateRequest)

		assert.NoError(t, err, "expected no error but got %v", err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), "expected status code to be %d but got %d", http.StatusOK, resp.StatusCode())
		assert.NotNil(t, enrollmentsResponse, "expected enrollments to not be nil")
		assert.Equal(t, status, *responseData.Status, "expected enrollment Status to be '%s' but got '%s'", status, *responseData.Status)

	})
}
