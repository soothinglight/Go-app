# Go + HTMX Authentication App

A modern web application demonstrating secure authentication and dynamic UI updates using Go and HTMX with an attractive, animated interface.

## ✨ Features

- 🔐 **Secure Authentication**: Cookie-based session management with bcrypt password hashing
- 🚀 **Dynamic UI**: Real-time updates with HTMX (no custom JavaScript required)
- 🎨 **Beautiful Design**: Animated login page with gradient backgrounds and glass morphism effects
- 📊 **Item Management**: Full CRUD operations with search functionality
- 🔍 **Live Search**: Real-time item filtering as you type
- 📱 **Responsive Design**: Works perfectly on desktop and mobile devices
- 🗄️ **SQLite Database**: Lightweight database with GORM ORM
- ✨ **Form Validation**: Inline validation with smooth error animations
- 🛡️ **Security First**: XSS protection, secure sessions, and proper authentication checks

## Prerequisites

- Go 1.21 or higher
- No additional dependencies required (Go modules will handle everything)

## How to Run

1. **Install dependencies and run:**
   ```bash
   go mod tidy && go run .
   ```

2. **Access the application:**
   - Open your browser to: http://localhost:8082
   - Login with seeded credentials: `admin@example.com` / `Passw0rd!`

## Architecture

### Routes
- `GET /` - Home page (login or dashboard based on auth status)
- `POST /login` - Authenticate user and return dashboard partial
- `POST /logout` - Destroy session and return login partial  
- `GET /items` - Get user's items list with optional search (authenticated)
- `POST /items` - Create new item and return updated list (authenticated)
- `DELETE /items/{id}` - Delete specific item and return updated list (authenticated)
- `GET /stats` - Get dashboard statistics (authenticated)

### Templates
- `base.templ` - Main layout with responsive design and login centering
- `login.templ` - Animated login form with gradient styling and glass morphism
- `dashboard.templ` - Clean dashboard with add item form and search functionality
- `items.templ` - Interactive items table with delete functionality

### Database Schema
```sql
-- Users table
users: id (pk), email (unique), password_hash, created_at

-- Items table  
items: id (pk), user_id (fk), name, created_at
```

### Security Features
- Passwords hashed with bcrypt
- Session cookies marked `HttpOnly` and `SameSite=Lax`
- Template XSS protection via `html/template`
- Server-side session validation on protected routes

## 🔄 HTMX Behavior

- **Login**: Form submits via HTMX with animated error handling
- **Logout**: Button triggers HTMX POST, smoothly returns to login view
- **Add Item**: Form submits and updates only the items table section
- **Search Items**: Live search with 300ms debounce for optimal performance
- **Delete Item**: Confirmation dialog with instant table updates
- **Load Items**: Items table lazy-loads on dashboard access
- **Error Handling**: All errors return styled HTML fragments with animations

## 📁 File Structure

```
├── main.go              # Main application with all handlers and models
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
├── templates/           # Template files (.templ extension)
│   ├── base.templ       # Main layout with responsive design
│   ├── login.templ      # Animated login form
│   ├── dashboard.templ  # Dashboard with search functionality
│   └── items.templ      # Interactive items table
├── app.db               # SQLite database (auto-created)
├── .gitignore           # Git ignore rules
└── README.md            # This documentation
```

## 🎬 Demo Flow

1. **Visit `/`** → Beautiful animated login page appears
2. **Wrong credentials** → Shake animation with error message, email preserved
3. **Correct credentials** → Smooth transition to dashboard without page refresh
4. **Add empty item** → Inline validation error in items table
5. **Add valid item** → Items table updates instantly with new entry
6. **Search items** → Real-time filtering as you type
7. **Delete item** → Confirmation dialog, then instant table update
8. **Logout** → Smooth transition back to animated login
9. **Unauthorized access** → Clean 401 HTML fragment response

## 🎨 UI Features

- **Gradient Backgrounds**: Beautiful purple-blue gradients
- **Glass Morphism**: Semi-transparent cards with backdrop blur
- **Smooth Animations**: Slide-up login, shake errors, hover effects
- **Interactive Elements**: Buttons lift on hover, inputs focus smoothly
- **Professional Typography**: Clean fonts with proper spacing
- **Responsive Design**: Adapts perfectly to all screen sizes

## 🛠️ Technical Details

### Development Notes
- **Zero Custom JavaScript**: Only HTMX script for all interactivity
- **Template-Based**: All responses return HTML partials for seamless updates
- **Auto-Migration**: Database schema updates automatically on startup
- **Seeded Data**: Admin user created automatically on first run
- **Session Management**: 7-day session expiration with secure cookies
- **Search Optimization**: Debounced search with SQL LIKE queries
- **Error Handling**: Graceful error responses with user-friendly messages

### Performance Features
- **Lazy Loading**: Items load only when dashboard is accessed
- **Debounced Search**: 300ms delay prevents excessive server requests
- **Efficient Queries**: Indexed user_id for fast item lookups
- **Minimal Payload**: Only necessary HTML fragments are transferred
- **CSS Animations**: Hardware-accelerated transforms for smooth effects

### Security Implementation
- **bcrypt Hashing**: Industry-standard password protection
- **Session Validation**: Every protected route checks authentication
- **XSS Prevention**: Go's html/template provides automatic escaping
- **CSRF Protection**: Session-based authentication prevents CSRF attacks
- **Input Validation**: Both client-side and server-side validation
- **Secure Headers**: HttpOnly and SameSite cookie flags