// Se indica que es el archivo principal
// Con go mod init <nombre modulo> permite crear un modulo
// Se instala el paquete para trabajar con postgresql con la instrucción: go get github.com/lib/pq
package main

import (
	"database/sql"  // objeto de tipo sql
	"log"           // Para mostrar en consola la información
	"net/http"      // Servidor http
	"text/template" // Modulo para trabajar con plantillas

	_ "github.com/lib/pq" // Se importa el driver de postgresql para usarse
)

// Función para conectarse a la base de datos
func conexionBD() (conexion *sql.DB) {

	// Variables para la conexión
	driver := "postgres"
	host := "localhost"
	user := "postgres"
	password := "postgres"
	dbname := "sistemago"

	// Intenta realizar la conexión, si se presenta error, se almacena en err.
	conexion, err := sql.Open(driver, "postgres://"+user+":"+password+"@"+host+"/"+dbname+"?sslmode=disable")

	if err != nil { // Si se presenta un error se detiene la goroutine y se muestra el error
		panic(err.Error())
	}
	return conexion // Si no hay error se retorna la conexion
}

var plantillas = template.Must(template.ParseGlob("plantillas/*")) // Se indica que se usaran todas las plantillas que estan en la ruta "plantillas/"

// Estructura de tipo empleado para leer información de los registros asociados a empleados
type Empleado struct {
	Id      int
	Nombre  string
	Correo  string
	Celular string
}

// Función principal
// Para correr la aplicación: go run main.go
func main() {

	http.HandleFunc("/", Start) // indica la ruta raiz y la función asociada que la maneja

	http.HandleFunc("/crear", Add) // Función para llevar al formulario de crear un nuevo empleado

	http.HandleFunc("/insertar", Insert) // Función para insertar un nuevo empleado

	http.HandleFunc("/borrar", Delete) // Función para borrar un empleado

	http.HandleFunc("/editar", Edit) // Función para llevar al formulario de edición de un empleado

	http.HandleFunc("/actualizar", Update) // Función para actualizar la informaciónd e un empleado

	log.Println("Servidor corriendo") // Muestra en consola un mensaje, se puede usar también fmt. Log indica hora del mensaje, fmt no.
	http.ListenAndServe(":9000", nil) // Se indica el puerto a usar, en este caso. 9000
}

// Función inicio recibe un parametro de tipo ResponseWriter para responder y otro de tipo http.request que representa la petición
// El *http.Request indica que es un puntero por lo que se almacenara en la variable será una dirección de memoria
func Start(w http.ResponseWriter, r *http.Request) {
	//Se imprime, con el writer, un mensaje
	//fmt.Fprintf(w, "Hola prueba")

	conexionEstablecida := conexionBD()

	registros, err := conexionEstablecida.Query("SELECT * from empleados")
	if err != nil {
		panic(err.Error())
	}
	empleado := Empleado{}          // Un empleado de tipo empleado
	arregloEmpleado := []Empleado{} // Un arreglo de empleados

	// Se recorren los registros
	for registros.Next() {
		var id int
		var nombre, correo, celular string // Variables

		err = registros.Scan(&id, &nombre, &correo, &celular) // Asigna en las variables los valores leidos de los registros, si hay error muestra el error
		if err != nil {
			panic(err.Error())
		}
		// Se almacenan los valores leidos en la variable empleado
		empleado.Id = id
		empleado.Nombre = nombre
		empleado.Correo = correo
		empleado.Celular = celular

		arregloEmpleado = append(arregloEmpleado, empleado) // Se agrega el empleado al arreglo de empleados
	}
	// fmt.Println(arregloEmpleado)

	plantillas.ExecuteTemplate(w, "inicio", arregloEmpleado) // Se indica que se ejecutara el template inicio pasandole el writer y el arreglo de empleados

}

// Función para agregar un empleado
func Add(w http.ResponseWriter, r *http.Request) {

	plantillas.ExecuteTemplate(w, "crear", nil) // Se indica que se ejecutara el template inicio pasandole el writer y sin parametros(nil)

}

// Función para almacenar los datos del empleado
func Insert(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		nombre := r.FormValue("nombre")
		correo := r.FormValue("correo")
		celular := r.FormValue("celular")

		conexionEstablecida := conexionBD()
		insertarRegistros, err := conexionEstablecida.Prepare("INSERT INTO empleados(nombre, correo, celular) VALUES ($1,$2,$3)") // La consulta para la inserción (varia lo del value según SGBD)

		if err != nil {

			panic(err.Error())
		} // Si hay error se muestra
		insertarRegistros.Exec(nombre, correo, celular) // Si no hay error se ejecuta la consulta con los valores recepcionados
		http.Redirect(w, r, "/", 301)                   // Redirecciona a la página principal con un código 301(Moved permanently)

	} // Si hay datos post guarda
}

// Función para borrar un empleado
func Delete(w http.ResponseWriter, r *http.Request) {
	idEmpleado := r.URL.Query().Get("id") // Recibo la id

	conexionEstablecida := conexionBD()
	borrarRegistro, err := conexionEstablecida.Prepare("DELETE FROM empleados WHERE id=$1") // La consulta para el borrado (varia lo del value según SGBD)

	if err != nil {
		panic(err.Error())
	} // Si hay error se muestra
	borrarRegistro.Exec(idEmpleado) // Si no hay error se ejecuta la consulta
	http.Redirect(w, r, "/", 301)   // Redirecciona a la página principal con un código 301(Moved permanently)

}

// Función para editar un empleado
func Edit(w http.ResponseWriter, r *http.Request) {

	idEmpleado := r.URL.Query().Get("id") // Recibo la id
	conexionEstablecida := conexionBD()

	registro, err := conexionEstablecida.Query("SELECT * from empleados WHERE id=$1", idEmpleado) // Se puede pasar el parametro en la consulta

	empleado := Empleado{} // Un empleado de tipo empleado

	// Se recorren los registros
	for registro.Next() {
		var id int
		var nombre, correo, celular string // Variables

		err = registro.Scan(&id, &nombre, &correo, &celular) // Asigna en las variables los valores leidos del registro si hay error muestra el error
		if err != nil {
			panic(err.Error())
		}
		// Se almacenan los valores leidos en la variable empleado
		empleado.Id = id
		empleado.Nombre = nombre
		empleado.Correo = correo
		empleado.Celular = celular
	}
	plantillas.ExecuteTemplate(w, "editar", empleado) // Se indica que se ejecutara el template editar con el empleado
}

// Función para actualizar los datos de un empleado
func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		id := r.FormValue("id")
		nombre := r.FormValue("nombre")
		correo := r.FormValue("correo")
		celular := r.FormValue("celular")

		conexionEstablecida := conexionBD()
		modificarRegistros, err := conexionEstablecida.Prepare("UPDATE empleados SET nombre=$1,correo=$2,celular=$3 WHERE id=$4") // La consulta para la actualización (varia lo del value según SGBD)

		if err != nil {

			panic(err.Error())
		} // Si hay error se muestra
		modificarRegistros.Exec(nombre, correo, celular, id) // Si no hay error se ejecuta la consulta con los valores recepcionados
		http.Redirect(w, r, "/", 301)                        // Redirecciona a la página principal con un código 301(Moved permanently)

	} // Si hay datos post guarda
}
