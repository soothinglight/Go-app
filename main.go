package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Models
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
}

type Item struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time
	User      User      `gorm:"foreignKey:UserID"`
}

// Global variables
var (
	db    *gorm.DB
	store *sessions.CookieStore
	tmpl  *template.Template
)

func main() {
	// Initialize database
	initDB()
	
	// Initialize session store
	store = sessions.NewCookieStore([]byte("your-secret-key-change-in-production"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	
	// Parse templates with custom functions
	var err error
	funcMap := template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
	tmpl = template.New("").Funcs(funcMap)
	tmpl, err = tmpl.ParseGlob("templates/*.templ")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}
	
	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/logout", logoutHandler).Methods("POST")
	r.HandleFunc("/items", itemsHandler).Methods("GET")
	r.HandleFunc("/items", createItemHandler).Methods("POST")
	r.HandleFunc("/items/{id}", deleteItemHandler).Methods("DELETE")
	r.HandleFunc("/stats", statsHandler).Methods("GET")
	
	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	
	fmt.Println("Server starting on http://localhost:8082")
	fmt.Println("Login with: admin@example.com / Passw0rd!")
	log.Fatal(http.ListenAndServe(":8082", r))
}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Auto migrate
	db.AutoMigrate(&User{}, &Item{})
	
	// Seed admin user if not exists
	var user User
	result := db.Where("email = ?", "admin@example.com").First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Passw0rd!"), bcrypt.DefaultCost)
		adminUser := User{
			Email:        "admin@example.com",
			PasswordHash: string(hashedPassword),
			CreatedAt:    time.Now(),
		}
		db.Create(&adminUser)
		fmt.Println("Admin user created: admin@example.com / Passw0rd!")
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	
	if ok && userID != nil {
		// User is logged in, show dashboard
		var user User
		db.First(&user, userID)
		data := map[string]interface{}{
			"User": user,
		}
		tmpl.ExecuteTemplate(w, "base.templ", map[string]interface{}{
			"Content": "dashboard",
			"Data":    data,
		})
	} else {
		// User not logged in, show login
		tmpl.ExecuteTemplate(w, "base.templ", map[string]interface{}{
			"Content": "login",
			"Data":    map[string]interface{}{},
		})
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	
	var user User
	result := db.Where("email = ?", email).First(&user)
	
	if result.Error != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		// Login failed - return login partial with error
		data := map[string]interface{}{
			"Error": "Invalid email or password",
			"Email": email,
		}
		tmpl.ExecuteTemplate(w, "login.templ", data)
		return
	}
	
	// Login successful - create session and return dashboard
	session, _ := store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Save(r, w)
	
	data := map[string]interface{}{
		"User": user,
	}
	tmpl.ExecuteTemplate(w, "dashboard.templ", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["user_id"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)
	
	// Return login partial
	tmpl.ExecuteTemplate(w, "login.templ", map[string]interface{}{})
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div class="error">Unauthorized. Please log in.</div>`))
		return
	}
	
	// Get search parameter
	search := r.URL.Query().Get("search")
	
	// Get user's items with optional search
	var items []Item
	query := db.Where("user_id = ?", userID)
	
	if search != "" {
		query = query.Where("name LIKE ? OR id LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	query.Order("created_at desc").Find(&items)
	
	data := map[string]interface{}{
		"Items": items,
	}
	tmpl.ExecuteTemplate(w, "items.templ", data)
}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div class="error">Unauthorized. Please log in.</div>`))
		return
	}
	
	name := r.FormValue("name")
	if name == "" {
		// Return error in items list format
		var items []Item
		db.Where("user_id = ?", userID).Order("created_at desc").Find(&items)
		data := map[string]interface{}{
			"Items": items,
			"Error": "Item name cannot be empty",
		}
		tmpl.ExecuteTemplate(w, "items.templ", data)
		return
	}
	
	// Convert userID to uint
	uid, _ := strconv.ParseUint(fmt.Sprintf("%v", userID), 10, 32)
	
	// Create item
	item := Item{
		UserID:    uint(uid),
		Name:      name,
		CreatedAt: time.Now(),
	}
	db.Create(&item)
	
	// Return updated items list
	var items []Item
	db.Where("user_id = ?", userID).Order("created_at desc").Find(&items)
	
	data := map[string]interface{}{
		"Items": items,
	}
	tmpl.ExecuteTemplate(w, "items.templ", data)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div class="error">Unauthorized. Please log in.</div>`))
		return
	}
	
	// Get item ID from URL
	vars := mux.Vars(r)
	itemID := vars["id"]
	
	// Delete item (only if it belongs to the user)
	db.Where("id = ? AND user_id = ?", itemID, userID).Delete(&Item{})
	
	// Return updated items list
	var items []Item
	db.Where("user_id = ?", userID).Order("created_at desc").Find(&items)
	
	data := map[string]interface{}{
		"Items": items,
	}
	tmpl.ExecuteTemplate(w, "items.templ", data)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div class="error">Unauthorized. Please log in.</div>`))
		return
	}
	
	// Get total items count
	var totalItems int64
	db.Model(&Item{}).Where("user_id = ?", userID).Count(&totalItems)
	
	// Get today's items count
	today := time.Now().Format("2006-01-02")
	var todayItems int64
	db.Model(&Item{}).Where("user_id = ? AND DATE(created_at) = ?", userID, today).Count(&todayItems)
	
	// Return stats as HTML fragment
	statsHTML := fmt.Sprintf(`
		<script>
			document.getElementById('total-items').textContent = '%d';
			document.getElementById('added-today').textContent = '%d';
			document.getElementById('items-count').textContent = '%d Total Items';
		</script>
	`, totalItems, todayItems, totalItems)
	
	w.Write([]byte(statsHTML))
}