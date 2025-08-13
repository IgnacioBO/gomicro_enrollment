// Aqui levantaremos un server (conexion bbdd, etc)
package test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"github.com/IgnacioBO/gomicro_enrollment/internal/enrollment"

	"github.com/IgnacioBO/gomicro_enrollment/pkg/bootstrap"
	"github.com/IgnacioBO/gomicro_enrollment/pkg/handler" //Manejar ruteo facilmente (paths y metodos)
	"github.com/joho/godotenv"

	courseSdkMock "github.com/IgnacioBO/go_micro_sdk/course/mock"
	userSdkMock "github.com/IgnacioBO/go_micro_sdk/user/mock"

	"github.com/IgnacioBO/go_http_client/client"
)

// Esta variables usaremos para hacer las peticioens http
// porque cuando levantemos el server hay que pegarles para ejecutar las peticoines
// Usaremos el client que craemo en el paquete httpclient
var cli client.Transport

// Primero copiearmoes el codigo (main dentro de TestMain y las otras funcioens aprte) de cmd/main.go  y lo copieamores debajo de TestMain por ahora.
func TestMain(m *testing.M) {

	//Aca usaremos el log con io.Discard para no mostrar logs en test
	l := log.New(io.Discard, "", 0)

	//Cargamos variables desde ../.env
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal(err)
	}

	//Creamos un config que tenga la cant max de pagina por defecto
	pageLimDef := os.Getenv("PAGINATOR_LIMIT_DEFAULT")
	if pageLimDef == "" {
		l.Fatal("paginator limit default is required")
	}
	enrollmentConfig := enrollment.Config{LimitPageDefault: pageLimDef}

	db, err := bootstrap.DBConnection()
	if err != nil {
		log.Fatal(err)
	}

	//********Para eviar tocar la bbdd que usamos normalmente,
	//EN vez de trabjar con la base de datos y pasarse al repo,
	//Generaremos una transaccion, y le pasaremos esa transaccion al repo
	//Y al finalizar LOS TEST HAREMOS UN ROLLBACK*****
	tx := db.Begin()

	//**Aqui pasaremos los SDK mockeadsos*
	userSdk := &userSdkMock.UserSdkMock{
		GetMock: func(id string) (*domain.User, error) {
			return &domain.User{}, nil
		},
	}

	courseSdk := &courseSdkMock.CourseSdkMock{
		GetMock: func(id string) (*domain.Course, error) {

			return &domain.Course{}, nil
		},
	}

	ctx := context.Background()
	enrollmentRepo := enrollment.NewRepo(l, tx)
	enrollmentService := enrollment.NewService(l, userSdk, courseSdk, enrollmentRepo)
	enrollmentEndpoint := enrollment.MakeEndpoints(enrollmentService, enrollmentConfig)
	h := handler.NewUserHTTPServer(ctx, enrollmentEndpoint)

	port := os.Getenv("PORT")
	address := fmt.Sprintf("127.0.0.1:%s", port)

	//**Aqui a la varaibles clin que generemos, le pasaremos el addres pero con http://**
	//Aqui usamle el client http que creamos (que iera lpara los sdk incialmente)
	cli = client.New(nil, "http://"+address, 0, false)

	srv := &http.Server{
		Handler:      accessControl(h), //Aquie le ponemos el acces control para deifnior op permitdas + el handler que definimos
		Addr:         address,
		ReadTimeout:  5 * time.Second, //Con estos SETEAMOS TIMEOUT DE ESCRITURA Y DE LECTURA (cuanto timepo maximo la api permite)
		WriteTimeout: 5 * time.Second, // Read es REQUEST, WRITE es RESPONE
	}
	//Ahora definim un channel, depsues de generar el server
	//Crearmos un goroutine, la ventaja es que podemos hacer otras cosas mientras se jecuta el serviodor ListenAndServe (como capturar en otro channel señales del sistema como conrtl+c u otras)
	errCh := make(chan error)

	//Aqui generamos una fncion anonimo de tipo GOROUTINE (por eso es **go func()**). Y la ejecutams altiro (por eso temrina en "()" despies de la llave "}" )
	//Ejecutamos el listenandserve() y si hay error retornamos eror al CANAL
	go func() {
		l.Println("listen in", address)
		errCh <- srv.ListenAndServe() //Aqui ejecutamos el ListenAndServe y VAMOS A DEVOLVER UN ERROR al CANAL
	}()

	r := m.Run()

	//**Aqui bajamos el servidor una vez trerminado las pruebas
	//Validamos si hay error**
	if err := srv.Shutdown(ctx); err != nil {
		log.Println((err))
	}

	// Y al finalizar LOS TEST HAREMOS UN ROLLBACK y eliminanos todos lso dartos de las pruebas, asi dejamos limpia la bbdd
	tx.Rollback()
	os.Exit(r)

}

// Aqui definimo operaciones que PERMITIREMOS y ademas recibimso un Handler original (que sera el que creamo en el main)
func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Aqui definimos operacione spermitidas, origin con * para que puedan venir DEDE CUALQUIER CLIENTE O LADO
		w.Header().Set("Access-Control-Allow-Origin", "*")                                //origin con * para que puedan venir DEDE CUALQUIER CLIENTE O LADO
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS, HEAD") //Metodos permitidos
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Requested-With") //Header permitidos

		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)

	})

}
