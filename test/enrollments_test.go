// Aqui por ejemplo seran los test del dominio enrollmentes, si hubieran otro dominio se crearia otro archivo de test
package test

import (
	"net/http"
	"testing"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"github.com/IgnacioBO/gomicro_enrollment/internal/enrollment"
	"github.com/stretchr/testify/assert"

	"github.com/IgnacioBO/go_lib_response/response"
)

//haremos un test rapido

// TEST FUCNIONALES

// aca generamoes una struct lde tipo dataRespone
func TestEnrollment(t *testing.T) {
	// Aqui ira la logica de nuestro test
	t.Run("should create an enrollment and get it", func(t *testing.T) {
		// Aqui ira la logica de nuestro test
		userid := "user1_test"
		courseid := "course1_test"
		bodyRequest := enrollment.CreateRequest{
			UserID:   userid,
			CourseID: courseid,
		}

		//Aqui usaremos el cliente [cli] (creados en main_test, que usa la librear http_client para pegarle a apis) y haermos un Post
		resp := cli.Post("/enrollments", bodyRequest)
		assert.Nil(t, resp.Err, "should not return an error")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "should return status code 201")

		dataCreated := domain.Enrollment{}
		//Esta variable tendra un struct de tipo SuccessResponse de la libreria go_lib_response (que igual podria crearlo nosotros un struct local si quieremos)
		//Aqui estamos poniendo Data: &dataCreated, le ponemos & como referencia para que se guarde la direccion de memoria
		//Es edecir que cuando se modifique dataCreated, se vera reflejado en dataRespCreated y viceversa
		dataRespCreated := response.SuccessResponse{Data: &dataCreated}

		//El respnse que recibimos
		//Aqui usamos FillUp, que lo que hace es llenar el struct de respuesta con los datos del BODY del response original
		//Osea lo que hace es tomar un struct y llenarlo con los datos de la respuesta (resp*)
		//Esto funcion porque recordad que el json de repsuesta es asi {status: x, data: {...}}
		//Y response.SuccessResponse tiene esa misma estructura gracias tambien a los tags json
		//Le pasamos tambien usando & para mantener la referencia, osea que entre en la funcion y permita modificar el original
		err := resp.FillUp(&dataRespCreated)
		assert.Nil(t, err, "should not return an error")

		assert.Equal(t, "success", dataRespCreated.Message, "should return status success")
		assert.Equal(t, http.StatusCreated, dataRespCreated.Status, "should return status code 201")
		assert.Equal(t, userid, dataCreated.UserID, "should return the same user id")
		assert.Equal(t, courseid, dataCreated.CourseID, "should return the same course id")
		assert.Equal(t, domain.Pending, dataCreated.Status, "should return status active")

		//Hacemos un get usando filtros user_id y course_id
		resp = cli.Get("/enrollments?user_id=" + dataCreated.UserID + "&course_id=" + dataCreated.CourseID)
		assert.Nil(t, resp.Err, "should not return an error")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "should return status code 200")

		dataGetAll := []domain.Enrollment{}
		dataRespGetAll := response.SuccessResponse{Data: &dataGetAll}

		err = resp.FillUp(&dataRespGetAll)
		assert.Nil(t, err, "should not return an error")

		assert.Equal(t, "success", dataRespGetAll.Message, "should return status success")
		assert.Equal(t, http.StatusOK, dataRespGetAll.Status, "should return status code 200")
		assert.Equal(t, dataCreated.ID, dataGetAll[0].ID, "should return the same enrollment id")
		assert.Equal(t, dataCreated.UserID, dataGetAll[0].UserID, "should return the same user id")
		assert.Equal(t, dataCreated.CourseID, dataGetAll[0].CourseID, "should return the same course id")
		assert.Equal(t, dataCreated.Status, dataGetAll[0].Status, "should return status active")
	})

	t.Run("should update an enrollment", func(t *testing.T) {
		//Primero creamos un create con un enrollmente
		userid := "user2_test"
		courseid := "course2_test"
		bodyRequest := enrollment.CreateRequest{
			UserID:   userid,
			CourseID: courseid,
		}

		resp := cli.Post("/enrollments", bodyRequest)
		assert.Nil(t, resp.Err, "should not return an error")
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "should return status code 201")

		dataCreated := domain.Enrollment{}
		dataRespCreated := response.SuccessResponse{Data: &dataCreated}
		err := resp.FillUp(&dataRespCreated)
		assert.Nil(t, err, "should not return an error")

		assert.Equal(t, "success", dataRespCreated.Message, "should return status success")

		//Ahora **actualizamos** el status
		status := "A"
		bodyRequestUpdate := enrollment.UpdateRequest{
			ID:     dataCreated.ID,
			Status: &status, //Cambiamos el status a active
		}

		resp = cli.Patch("/enrollments/"+dataCreated.ID, bodyRequestUpdate)
		assert.Nil(t, resp.Err, "should not return an error")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "should return status code 200")

		dataUpdated := domain.Enrollment{}
		dataRespUpdated := response.SuccessResponse{Data: &dataUpdated}
		err = resp.FillUp(&dataRespUpdated)
		assert.Nil(t, err, "should not return an error")

		assert.Equal(t, "success", dataRespUpdated.Message, "should return status success")
		assert.Equal(t, http.StatusOK, dataRespUpdated.Status, "should return status code 200")
		assert.Equal(t, domain.Active, dataUpdated.Status, "should return status active")

		//Ahora un **getall**
		//Hacemos un get usando filtros user_id y course_id
		resp = cli.Get("/enrollments?user_id=" + dataCreated.UserID + "&course_id=" + dataCreated.CourseID)
		assert.Nil(t, resp.Err, "should not return an error")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "should return status code 200")

		dataGetAll := []domain.Enrollment{}
		dataRespGetAll := response.SuccessResponse{Data: &dataGetAll}

		err = resp.FillUp(&dataRespGetAll)
		assert.Nil(t, err, "should not return an error")

		assert.Equal(t, "success", dataRespGetAll.Message, "should return status success")
		assert.Equal(t, http.StatusOK, dataRespGetAll.Status, "should return status code 200")
		assert.Equal(t, dataUpdated.ID, dataGetAll[0].ID, "should return the same enrollment id")
		assert.Equal(t, dataUpdated.Status, dataGetAll[0].Status, "should return status active")

	})
}
