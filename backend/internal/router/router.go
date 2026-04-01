package router

import (
	"net/http"

	"autoservice/backend/internal/api"
	"autoservice/backend/internal/auth"
	"autoservice/backend/internal/cars"
	"autoservice/backend/internal/config"
	"autoservice/backend/internal/employees"
	"autoservice/backend/internal/handlers"
	"autoservice/backend/internal/maintenanceparts"
	"autoservice/backend/internal/orders"
	"autoservice/backend/internal/owners"
	"autoservice/backend/internal/repairparts"
	"autoservice/backend/internal/services"
	"autoservice/backend/internal/users"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, db *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	homeHandler := handlers.NewHomeHandler(cfg.TemplatesDir)
	healthHandler := handlers.NewHealthHandler(db)
	sessionManager := auth.NewSessionManager(cfg)

	ownersRepo := owners.NewRepository(db)
	ownersHandler := owners.NewHandler(ownersRepo)

	carsRepo := cars.NewRepository(db)
	carsHandler := cars.NewHandler(carsRepo)

	servicesRepo := services.NewRepository(db)
	servicesHandler := services.NewHandler(servicesRepo)

	employeesRepo := employees.NewRepository(db)
	employeesHandler := employees.NewHandler(employeesRepo)

	repairPartsRepo := repairparts.NewRepository(db)
	repairPartsHandler := repairparts.NewHandler(repairPartsRepo)

	maintenancePartsRepo := maintenanceparts.NewRepository(db)
	maintenancePartsHandler := maintenanceparts.NewHandler(maintenancePartsRepo)

	ordersRepo := orders.NewRepository(db)
	ordersHandler := orders.NewHandler(ordersRepo)

	usersRepo := users.NewRepository(db)
	usersHandler := users.NewHandler(usersRepo, sessionManager)

	staticFS := http.FileServer(http.Dir(cfg.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	mux.HandleFunc("/", homeHandler.Index)
	mux.HandleFunc("/api/health", healthHandler.Check)

	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			usersHandler.Register(w, r)
		default:
			api.MethodNotAllowed(w, "POST")
		}
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			usersHandler.Login(w, r)
		default:
			api.MethodNotAllowed(w, "POST")
		}
	})

	mux.HandleFunc("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			usersHandler.Logout(w, r)
		default:
			api.MethodNotAllowed(w, "POST")
		}
	})

	mux.HandleFunc("/api/profile/me", sessionManager.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			usersHandler.GetCurrentProfile(w, r)
		case http.MethodPut:
			usersHandler.UpdateCurrentProfile(w, r)
		default:
			api.MethodNotAllowed(w, "GET, PUT")
		}
	}))

	mux.HandleFunc("/api/owners", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ownersHandler.GetAll(w, r)
		case http.MethodPost:
			ownersHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/owners/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			ownersHandler.Update(w, r)
		case http.MethodDelete:
			ownersHandler.Delete(w, r)
		default:
			api.MethodNotAllowed(w, "PUT, DELETE")
		}
	})

	mux.HandleFunc("/api/cars", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			carsHandler.GetAll(w, r)
		case http.MethodPost:
			carsHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/cars/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			carsHandler.Update(w, r)
		case http.MethodDelete:
			carsHandler.Delete(w, r)
		default:
			api.MethodNotAllowed(w, "PUT, DELETE")
		}
	})

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			servicesHandler.GetAll(w, r)
		case http.MethodPost:
			servicesHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/services/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			servicesHandler.Update(w, r)
		case http.MethodDelete:
			servicesHandler.Delete(w, r)
		default:
			api.MethodNotAllowed(w, "PUT, DELETE")
		}
	})

	mux.HandleFunc("/api/employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			employeesHandler.GetAll(w, r)
		case http.MethodPost:
			employeesHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			employeesHandler.Update(w, r)
		case http.MethodDelete:
			employeesHandler.Delete(w, r)
		default:
			api.MethodNotAllowed(w, "PUT, DELETE")
		}
	})

	mux.HandleFunc("/api/repair-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			repairPartsHandler.GetAll(w, r)
		case http.MethodPost:
			repairPartsHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/maintenance-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			maintenancePartsHandler.GetAll(w, r)
		case http.MethodPost:
			maintenancePartsHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.GetAll(w, r)
		case http.MethodPost:
			ordersHandler.Create(w, r)
		default:
			api.MethodNotAllowed(w, "GET, POST")
		}
	})

	mux.HandleFunc("/api/orders/form-data", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.GetFormData(w, r)
		default:
			api.MethodNotAllowed(w, "GET")
		}
	})

	mux.HandleFunc("/api/orders/details", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.GetDetails(w, r)
		default:
			api.MethodNotAllowed(w, "GET")
		}
	})

	mux.HandleFunc("/api/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.GetByID(w, r)
		case http.MethodPut:
			ordersHandler.Update(w, r)
		case http.MethodDelete:
			ordersHandler.Delete(w, r)
		default:
			api.MethodNotAllowed(w, "GET, PUT, DELETE")
		}
	})

	mux.HandleFunc("/api/orders/{id}/employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			ordersHandler.ReplaceEmployees(w, r)
		default:
			api.MethodNotAllowed(w, "PUT")
		}
	})

	mux.HandleFunc("/api/orders/{id}/repair-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			ordersHandler.ReplaceRepairParts(w, r)
		default:
			api.MethodNotAllowed(w, "PUT")
		}
	})

	mux.HandleFunc("/api/orders/{id}/maintenance-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			ordersHandler.ReplaceMaintenanceParts(w, r)
		default:
			api.MethodNotAllowed(w, "PUT")
		}
	})

	mux.HandleFunc("/api/orders/assign-employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ordersHandler.AssignEmployees(w, r)
		default:
			api.MethodNotAllowed(w, "POST")
		}
	})

	mux.HandleFunc("/api/orders/add-repair-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ordersHandler.AddRepairParts(w, r)
		default:
			api.MethodNotAllowed(w, "POST")
		}
	})

	mux.HandleFunc("/api/orders/add-maintenance-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ordersHandler.AddMaintenanceParts(w, r)
		default:
			api.MethodNotAllowed(w, "POST")
		}
	})

	return mux
}
