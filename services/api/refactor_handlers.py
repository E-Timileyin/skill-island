import os
import glob

api_dir = "/home/eyiowuawi/Documents/projects/skill-island/services/api/internal/api"
handlers_dir = os.path.join(api_dir, "handlers")
routes_dir = os.path.join(api_dir, "routes")

# Move other files to handlers
files_to_move = glob.glob(os.path.join(api_dir, "*.go"))
for f in files_to_move:
    filename = os.path.basename(f)
    if filename == "handler.go":
        continue # we will split this
        
    with open(f, "r") as file:
        content = file.read()
        
    content = content.replace("package api\n", "package handlers\n")
    
    new_path = os.path.join(handlers_dir, filename)
    with open(new_path, "w") as file:
        file.write(content)
        
    os.remove(f)

# Split handler.go
with open(os.path.join(api_dir, "handler.go"), "r") as f:
    handler_content = f.read()

base_go = """package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APIError is the standard error response format.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	DB  *pgxpool.Pool
	Cfg config.Config
}

// Health returns the service health status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
"""

auth_go = handler_content.replace(
    'package api', 'package handlers'
)

# Remove the parts that are now in base.go from auth.go
import re
auth_go = re.sub(r'// APIError.*?// Health.*?}\n+', '', auth_go, flags=re.DOTALL)
auth_go = re.sub(r'// writeJSON.*?}\n', '', auth_go, flags=re.DOTALL)

with open(os.path.join(handlers_dir, "base.go"), "w") as f:
    f.write(base_go)
    
with open(os.path.join(handlers_dir, "auth.go"), "w") as f:
    f.write(auth_go)

os.remove(os.path.join(api_dir, "handler.go"))
