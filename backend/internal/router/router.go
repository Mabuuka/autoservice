package router

import (
	"net/http"

	"kursovaya/backend/internal/cars"
	"kursovaya/backend/internal/config"
	"kursovaya/backend/internal/employees"
	"kursovaya/backend/internal/handlers"
	"kursovaya/backend/internal/maintenanceparts"
	"kursovaya/backend/internal/orders"
	"kursovaya/backend/internal/owners"
	"kursovaya/backend/internal/repairparts"
	"kursovaya/backend/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, db *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	homeHandler := handlers.NewHomeHandler(cfg.TemplatesDir)
	healthHandler := handlers.NewHealthHandler(db)

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

	staticFS := http.FileServer(http.Dir(cfg.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	mux.HandleFunc("/", homeHandler.Index)
	mux.HandleFunc("/api/health", healthHandler.Check)

	mux.HandleFunc("/api/owners", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ownersHandler.GetAll(w, r)
		case http.MethodPost:
			ownersHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/cars", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			carsHandler.GetAll(w, r)
		case http.MethodPost:
			carsHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			servicesHandler.GetAll(w, r)
		case http.MethodPost:
			servicesHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			employeesHandler.GetAll(w, r)
		case http.MethodPost:
			employeesHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/repair-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			repairPartsHandler.GetAll(w, r)
		case http.MethodPost:
			repairPartsHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/maintenance-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			maintenancePartsHandler.GetAll(w, r)
		case http.MethodPost:
			maintenancePartsHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.GetAll(w, r)
		case http.MethodPost:
			ordersHandler.Create(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/orders/details", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ordersHandler.GetDetails(w, r)
		default:
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/orders/assign-employees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ordersHandler.AssignEmployees(w, r)
		default:
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/orders/add-repair-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ordersHandler.AddRepairParts(w, r)
		default:
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/orders/add-maintenance-parts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ordersHandler.AddMaintenanceParts(w, r)
		default:
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
