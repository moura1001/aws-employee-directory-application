package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/moura1001/aws-employee-directory-application/server/model"
	"github.com/moura1001/aws-employee-directory-application/server/store"
	"github.com/moura1001/aws-employee-directory-application/server/utils"
)

type Server struct {
	store store.EmployeeStore
	http.Handler
	maxBytesReader int64
}

func NewServer() (*Server, error) {
	server := new(Server)
	if utils.DYNAMO_MODE != "" {
		server.store = store.NewDynamoStore()
	} else {
		server.store = store.NewInMemoryStore()
	}

	router := mux.NewRouter()
	router.HandleFunc("/", server.home).Methods("GET")
	router.HandleFunc("/add", server.add).Methods("GET")
	router.HandleFunc("/edit/{employeeId}", server.edit).Methods("GET")
	router.HandleFunc("/save", server.save).Methods("POST")
	router.HandleFunc("/employee/{employeeId}", server.view).Methods("GET")
	router.HandleFunc("/delete/{employeeId}", server.delete).Methods("GET")

	server.Handler = csrf.Protect(
		[]byte(utils.CSRF_SECRET),
		csrf.Path("/"),
	)(router)

	server.maxBytesReader = 1<<20 + 1024

	return server, nil
}

func urlFor(host string, endpoint string) string {
	return "http://" + host + endpoint
}

func (server *Server) home(w http.ResponseWriter, r *http.Request) {
	//s3_client = boto3.client('s3')
	employees := server.store.ListEmployees()
	if len(employees) == 0 {
		url := urlFor(r.Host, "/add")
		templateStr := fmt.Sprintf(`
			{{template "main" .}}
			{{define "head"}}
			Employee Directory - Home
			<a class="btn btn-primary float-right" href="%s">Add</a>
			{{end}}
			{{define "body"}}Empty{{end}}
		`, url)

		t, _ := template.New("home").Parse(templateStr)
		t, err := t.ParseFiles("./static/templates/main.html")
		if err == nil {
			err = t.Execute(w, nil)
			if err != nil {
				fmt.Fprintf(w, "error to execute template: %+v\n", err)
			}
		}
		return
	} else {
		for _, employee := range employees {
			if employee.Photo.ObjectKey != "" {
				/*employee.Photo.SignedUrl = s3_client.generate_presigned_url(
				    'get_object',
				    Params={'Bucket': config.PHOTOS_BUCKET, 'Key': employee["object_key"]}
				)*/
			}
		}
	}

	urlAdd := urlFor(r.Host, "/add")
	urlDelete := urlFor(r.Host, "/delete")
	urlView := urlFor(r.Host, "/employee")

	templateStr := fmt.Sprintf(`
	{{ template "main" .}}
	{{ define "head" }}
	Employee Directory - Home
	<a class="btn btn-primary float-right" href="%s">Add</a>
	{{ end }}
	{{ define "body" }}
		{{  if not .employees }}<h4>Empty Directory</h4>{{ end }}

		<table class="table table-bordered">
		  <tbody>
		  	{{ $badges := .badges }}
			{{ range $employee := .employees }}
				<tr>
				<td width="100">{{ if $employee.Photo.SignedUrl }}
				<img width="50" src="{{$employee.Photo.SignedUrl}}" /><br/>
				{{ end }}
				<a href="%s/{{$employee.Id}}"><span class="fa fa-remove" aria-hidden="true"></span> delete</a>
				</td>
				<td><a href="%s/{{$employee.Id}}">{{$employee.FullName}}</a>
				{{ range $key, $badge := $badges }}
				{{ if $employee.HasBadge $key }}
				<i class="fa fa-{{$key}}" title="{{$badge}}"></i>
				{{ end }}
				{{ end }}
				<br/>
				<small>{{$employee.Location}}</small>
				</td>
				</tr>
			{{ end }}

		  </tbody>
		</table>

	{{ end }}
	`, urlAdd, urlDelete, urlView)

	t, _ := template.New("home").Parse(templateStr)
	t, err := t.ParseFiles("./static/templates/main.html")
	if err == nil {
		err = t.Execute(w, map[string]interface{}{
			"employees": employees,
			"badges":    model.Badges,
		})
		if err != nil {
			fmt.Fprintf(w, "error to execute template: %+v\n", err)
		}
	}
}

func (server *Server) add(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/templates/view-edit.html", "./static/templates/main.html")
	if err == nil {
		err = t.Execute(w, map[string]interface{}{
			"form":       model.NewForm(),
			"badges":     model.Badges,
			"url_save":   urlFor(r.Host, "/save"),
			"csrf_token": csrf.Token(r),
		})
		if err != nil {
			fmt.Fprintf(w, "error to execute template: %+v\n", err)
		}
	}
}

func (server *Server) edit(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	//s3_client = boto3.client('s3')
	employee := server.store.LoadEmployee(params["employeeId"])
	signedUrl := ""
	if employee == nil {
		fmt.Fprintf(w, "employee not found")
		return
	}

	if employee.Photo.ObjectKey != "" {
		/*signed_url = s3_client.generate_presigned_url(
		    'get_object',
		    Params={'Bucket': config.PHOTOS_BUCKET, 'Key': employee["object_key"]}
		)*/
	}

	form := model.NewForm()
	form.EmployeeId.Data = employee.Id
	form.FullName.Data = employee.FullName
	form.Location.Data = employee.Location
	form.JobTitle.Data = employee.JobTitle
	if len(employee.Badges) > 0 {
		form.Badges.Data = employee.Badges
	}

	t, err := template.ParseFiles("./static/templates/view-edit.html", "./static/templates/main.html")
	if err == nil {
		err = t.Execute(w, map[string]interface{}{
			"form":       form,
			"badges":     model.Badges,
			"url_save":   urlFor(r.Host, "/save"),
			"signed_url": signedUrl,
			"csrf_token": csrf.Token(r),
		})
		if err != nil {
			fmt.Fprintf(w, "error to execute template: %+v\n", err)
		}
	}
}

func (server *Server) save(w http.ResponseWriter, r *http.Request) {

	r.Body = http.MaxBytesReader(w, r.Body, server.maxBytesReader)
	err := r.ParseMultipartForm(server.maxBytesReader)
	if err != nil {
		http.Error(w, fmt.Errorf("error to parse form data: %v", err).Error(), http.StatusBadRequest)
		return
	}

	form := model.NewForm()
	err = form.ValidateOnSubmit(r.MultipartForm)

	if err == nil {

		employeeId := form.EmployeeId.Data.(string)
		key := ""
		fullName := form.FullName.Data.(string)
		location := form.Location.Data.(string)
		jobTitle := form.JobTitle.Data.(string)
		badges := form.Badges.Data.([]string)

		if employeeId != "" {
			server.store.UpdateEmployee(
				employeeId,
				key,
				fullName,
				location,
				jobTitle,
				badges,
			)
		} else {
			server.store.AddEmployee(
				key,
				fullName,
				location,
				jobTitle,
				badges,
			)
		}
		//flash("Saved!")
		//return redirect(url_for("home"))

		http.Redirect(w, r, urlFor(r.Host, "/"), http.StatusMovedPermanently)
	} else {
		http.Error(w, fmt.Errorf("form failed validate: %v", err).Error(), http.StatusBadRequest)
	}

	/*"Save an employee"
	  form = EmployeeForm()
	  s3_client = boto3.client('s3')
	  key = None
	  if form.validate_on_submit():
	      if form.photo.data:
	          image_bytes = util.resize_image(form.photo.data, (120, 160))
	          if image_bytes:
	              try:
	                  # save the image to s3
	                  prefix = "employee_pic/"
	                  key = prefix + util.random_hex_bytes(8) + '.png'
	                  s3_client.put_object(
	                      Bucket=config.PHOTOS_BUCKET,
	                      Key=key,
	                      Body=image_bytes,
	                      ContentType='image/png'
	                  )
	              except:
	                  pass*/
}

func (server *Server) view(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	//s3_client = boto3.client('s3')
	employee := server.store.LoadEmployee(params["employeeId"])
	if employee == nil {
		fmt.Fprintf(w, "employee not found")
		return
	}

	if employee.Photo.ObjectKey != "" {
		/*employee["signed_url"] = s3_client.generate_presigned_url(
		    'get_object',
		    Params={'Bucket': config.PHOTOS_BUCKET, 'Key': employee["object_key"]}
		)*/
	}

	urlEdit := urlFor(r.Host, "/edit")
	urlHome := urlFor(r.Host, "/")

	templateStr := fmt.Sprintf(`
	    {{ template "main" .}}
	    {{ define "head" }}
	        {{.employee.FullName}}
	        <a class="btn btn-primary float-right" href="%s/{{.employee.Id}}">Edit</a>
	        <a class="btn btn-primary float-right" href="%s">Home</a>
	    {{ end }}
	    {{ define "body" }}

	  	<div class="row">
			<div class="col-md-4">
				{{ if .employee.Photo.SignedUrl }}
				<img alt="Mugshot" src="{{ .employee.Photo.SignedUrl }}" />
				{{ end }}
			</div>

	    	<div class="col-md-8">
				<div class="form-group row">
					<label class="col-sm-2">{{ .form.Location.Label }}</label>
					<div class="col-sm-10">
					{{.employee.Location}}
					</div>
				</div>
	      		<div class="form-group row">
	        		<label class="col-sm-2">{{.form.JobTitle.Label}}</label>
					<div class="col-sm-10">
					{{.employee.JobTitle}}
					</div>
	      		</div>
				{{ $employee := .employee }}
				{{ range $key, $badge := .badges }}
				<div class="form-check">
					{{ if $employee.HasBadge $key }}
					<span class="badge badge-primary"><i class="fa fa-{{$key}}"></i> {{$badge}}</span>
					{{ end }}
				</div>
				{{ end }}
				&nbsp;
	    	</div>
	  	</div>
	    {{ end }}
		`, urlEdit, urlHome)

	t, _ := template.New("view").Parse(templateStr)
	t, err := t.ParseFiles("./static/templates/main.html")
	if err == nil {
		err = t.Execute(w, map[string]interface{}{
			"form":     model.NewForm(),
			"badges":   model.Badges,
			"employee": employee,
		})
		if err != nil {
			fmt.Fprintf(w, "error to execute template: %+v\n", err)
		}
	}
}

func (server *Server) delete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	server.store.DeleteEmployee(params["employeeId"])
	//flash("Deleted!")
	http.Redirect(w, r, urlFor(r.Host, "/"), http.StatusMovedPermanently)
}
