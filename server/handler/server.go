package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/moura1001/aws-employee-directory-application/server/model"
	"github.com/moura1001/aws-employee-directory-application/server/store"
	"github.com/moura1001/aws-employee-directory-application/server/utils"
)

type Server struct {
	store    store.EmployeeStore
	s3Client store.S3Store
	http.Handler
	maxBytesReader   int64
	availabilityZone string
	instanceId       string
	session          *sessions.CookieStore
	sessionName      string
	flashTemplate    string
}

func NewServer() (*Server, error) {
	server := new(Server)

	server.session = sessions.NewCookieStore([]byte(utils.SESSION_KEY))
	server.sessionName = "employee-session"
	server.flashTemplate = "flashed_messages"

	if err := server.setInstanceDocumentInfo(); err != nil {
		log.Printf(" * Instance metadata not available. Details: '%s'\n", err)
		server.availabilityZone = "us-fake-1a"
		server.instanceId = "i-fakeabc"
	}

	if utils.DYNAMO_MODE == "on" {
		server.store = store.NewDynamoStore()
	} else {
		server.store = store.NewMysqlStore()
	}

	server.s3Client = store.NewS3Store()

	router := mux.NewRouter()
	router.HandleFunc("/", server.home).Methods("GET")
	router.HandleFunc("/add", server.add).Methods("GET")
	router.HandleFunc("/edit/{employeeId}", server.edit).Methods("GET")
	router.HandleFunc("/save", server.save).Methods("POST")
	router.HandleFunc("/employee/{employeeId}", server.view).Methods("GET")
	router.HandleFunc("/delete/{employeeId}", server.delete).Methods("GET")
	router.HandleFunc("/info", server.info).Methods("GET")
	router.HandleFunc("/info/stress_cpu/{seconds}", server.stress).Methods("GET")
	router.HandleFunc("/monitor", server.monitor).Methods("GET")

	server.Handler = csrf.Protect(
		[]byte(utils.CSRF_SECRET),
		csrf.Path("/"),
		csrf.Secure(false),
	)(router)

	server.maxBytesReader = 1<<20 + 1024

	return server, nil
}

func (server *Server) setInstanceDocumentInfo() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(opts *config.LoadOptions) error {
		opts.Region = utils.AWS_DEFAULT_REGION
		return nil
	})
	if err != nil {
		return fmt.Errorf("error to load default config. Details: '%s'", err)
	}

	client := imds.NewFromConfig(cfg)

	iido, err := client.GetInstanceIdentityDocument(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("error to get ec2 instance identity document. Details: '%s'", err)
	}

	server.availabilityZone = iido.AvailabilityZone
	server.instanceId = iido.InstanceID

	return nil
}

func urlFor(host string, endpoint string) string {
	return "http://" + host + endpoint
}

func (server *Server) home(w http.ResponseWriter, r *http.Request) {
	session, _ := server.session.Get(r, server.sessionName)
	flashedMessages, _ := session.Values[server.flashTemplate].([]string)
	if len(flashedMessages) > 0 {
		session.Values[server.flashTemplate] = nil
		session.Save(r, w)
	}

	employees, err := server.store.ListEmployees()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
			err = t.Execute(w, map[string]interface{}{
				server.flashTemplate: flashedMessages,
			})
			if err != nil {
				fmt.Fprintf(w, "error to execute template: %+v\n", err)
			}
		}
		return
	} else {
		for _, employee := range employees {
			if employee.Photo.ObjectKey != "" {
				url, err := server.s3Client.GeneratePresignedURL(employee.Photo.ObjectKey)
				if err == nil {
					employee.Photo.SignedUrl = url
				} else {
					employee.Photo.SignedUrl = err.Error()
				}
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
	t, err = t.ParseFiles("./static/templates/main.html")
	if err == nil {
		err = t.Execute(w, map[string]interface{}{
			"employees":          employees,
			"badges":             model.Badges,
			server.flashTemplate: flashedMessages,
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

	employee, err := server.store.LoadEmployee(params["employeeId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signedUrl := ""
	if employee == nil {
		http.Error(w, "employee not found", http.StatusNotFound)
		return
	}

	if employee.Photo.ObjectKey != "" {
		url, err := server.s3Client.GeneratePresignedURL(employee.Photo.ObjectKey)
		if err == nil {
			signedUrl = url
		} else {
			signedUrl = err.Error()
		}
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
	session, _ := server.session.Get(r, server.sessionName)

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
		fullName := form.FullName.Data.(string)
		location := form.Location.Data.(string)
		jobTitle := form.JobTitle.Data.(string)
		badges := form.Badges.Data.([]string)

		if employeeId == "" {
			employeeId, err = server.store.AddEmployee(
				"",
				fullName,
				location,
				jobTitle,
				badges,
			)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if employeeId != "" {
			key := ""
			if form.Photo.Data != nil {
				imageBytes, err := utils.ResizeImage(form.Photo.Data.([]byte), 120, 160)
				if err != nil {
					http.Error(w, fmt.Errorf("error to resize image: %v", err).Error(), http.StatusInternalServerError)
					return
				}

				// save the image to s3
				prefix := "employee_pic/"
				key = prefix + employeeId + ".png"
				err = server.s3Client.UploadObject(key, imageBytes)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			err = server.store.UpdateEmployee(
				employeeId,
				key,
				fullName,
				location,
				jobTitle,
				badges,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		session.Values[server.flashTemplate] = []string{"Saved!"}
		session.Save(r, w)
		//flash("Saved!")
		//return redirect(url_for("home"))

		http.Redirect(w, r, urlFor(r.Host, "/"), http.StatusMovedPermanently)
	} else {
		http.Error(w, fmt.Errorf("form failed validate: %v", err).Error(), http.StatusBadRequest)
	}
}

func (server *Server) view(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	employee, err := server.store.LoadEmployee(params["employeeId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if employee == nil {
		http.Error(w, "employee not found", http.StatusNotFound)
		return
	}

	if employee.Photo.ObjectKey != "" {
		url, err := server.s3Client.GeneratePresignedURL(employee.Photo.ObjectKey)
		if err == nil {
			employee.Photo.SignedUrl = url
		} else {
			employee.Photo.SignedUrl = err.Error()
		}
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
	t, err = t.ParseFiles("./static/templates/main.html")
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
	session, _ := server.session.Get(r, server.sessionName)

	params := mux.Vars(r)
	err := server.store.DeleteEmployee(params["employeeId"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values[server.flashTemplate] = []string{"Deleted!"}
	session.Save(r, w)
	//flash("Deleted!")

	http.Redirect(w, r, urlFor(r.Host, "/"), http.StatusMovedPermanently)
}

func (server *Server) info(w http.ResponseWriter, r *http.Request) {
	session, _ := server.session.Get(r, server.sessionName)

	urlStress60 := urlFor(r.Host, "/info/stress_cpu/60")
	urlStress300 := urlFor(r.Host, "/info/stress_cpu/300")
	urlStress600 := urlFor(r.Host, "/info/stress_cpu/600")

	templateStr := fmt.Sprintf(`
		{{ template "main" .}}
		{{ define "head" }}
			Instance Info
		{{ end }}
		{{ define "body" }}
		<b>instance_id</b>: {{.g.instance_id}} <br/>
		<b>availability_zone</b>: {{.g.availablity_zone}} <br/>
		<hr/>
		<small>Stress cpu:
		<a href="%s">1 min</a>,
		<a href="%s">5 min</a>,
		<a href="%s">10 min</a>
		</small>
		{{ end }}
	`, urlStress60, urlStress300, urlStress600)

	t, _ := template.New("info").Parse(templateStr)
	t, err := t.ParseFiles("./static/templates/main.html")
	if err == nil {
		flashedMessages, _ := session.Values[server.flashTemplate].([]string)
		if len(flashedMessages) > 0 {
			session.Values[server.flashTemplate] = nil
			session.Save(r, w)
		}

		err = t.Execute(w, map[string]interface{}{
			"g": map[string]string{
				"instance_id":      server.instanceId,
				"availablity_zone": server.availabilityZone,
			},
			server.flashTemplate: flashedMessages,
		})
		if err != nil {
			fmt.Fprintf(w, "error to execute template: %+v\n", err)
		}
	}
}

func (server *Server) stress(w http.ResponseWriter, r *http.Request) {
	session, _ := server.session.Get(r, server.sessionName)

	params := mux.Vars(r)
	if params["seconds"] == "60" || params["seconds"] == "300" || params["seconds"] == "600" {
		err := exec.Command("stress", "--cpu", "8", "--timeout", params["seconds"]).Start()
		if err != nil {
			msgErr := fmt.Errorf("error to simulate cpu stress with param '%s'. Details: '%s'", params["seconds"], err).Error()
			http.Error(w, msgErr, http.StatusInternalServerError)
			return
		}

		session.Values[server.flashTemplate] = []string{"Stressing CPU"}
		session.Save(r, w)
		//flash("Stressing CPU")

		http.Redirect(w, r, urlFor(r.Host, "/info"), http.StatusMovedPermanently)
	}
}

func (server *Server) monitor(w http.ResponseWriter, r *http.Request) {
	healthStatus := map[bool]string{true: "OK", false: "PROBLEM"}

	isDbHealthy := server.store.IsHealthy()
	isS3Healthy := server.s3Client.IsHealthy()

	msg := fmt.Sprintf("s3 status: %s\ndatabase status: %s\n", healthStatus[isDbHealthy], healthStatus[isS3Healthy])

	if isDbHealthy && isS3Healthy {
		w.Write([]byte(msg))
	} else {
		http.Error(w, msg, http.StatusServiceUnavailable)
	}
}
