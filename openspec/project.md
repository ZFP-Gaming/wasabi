# Project Context

## Purpose
Wasabi is a simple file upload and management API for audio files. The project provides a REST API backend in Go that handles file uploads, listing, renaming, deletion, and serving of audio files stored locally in an `uploads` directory. A React frontend provides a user interface for interacting with the API.

## Tech Stack

### Backend
- **Go 1.24.2**: Main backend language
- **Standard library HTTP server**: No external web frameworks
- **File-based storage**: Local filesystem for audio file storage

### Frontend
- **React 18.2**: UI framework
- **Vite 5.2**: Build tool and dev server
- **phosphor-react**: Icon library
- **react-mirt**: Additional React utilities

## Project Conventions

### Code Style
- **Go**: Standard Go conventions following `gofmt`
  - camelCase for local variables and struct fields
  - PascalCase for exported identifiers
  - Spanish language used in comments and error messages
  - Error handling with explicit checks and logging
  - Simple, readable code over clever abstractions

- **JavaScript/React**:
  - Modern ES6+ syntax
  - Functional components with hooks
  - Component-based architecture in `frontend/src/components/`
  - Service layer pattern for API calls in `frontend/src/services/`

### Architecture Patterns
- **Backend**: Simple REST API with three main handlers:
  - `uploadHandler`: POST /upload (multipart file upload)
  - `listHandler`: GET /files (list all files)
  - `fileHandler`: GET/PUT/DELETE /files/{name} (file operations)
  - Request logging middleware
  - Path sanitization for security

- **Frontend**: Component-based with separation of concerns
  - `components/`: UI components (FileList, UploadForm)
  - `services/`: API interaction layer (api.js, audio.js)
  - Centralized API base URL configuration

- **Configuration**: Environment-based via `.env` file
  - `PORT`: Server listen address (default: 8080)
  - `UPLOAD_DIR`: File storage directory (default: uploads)
  - Custom `.env` parser (no external dependencies)

### Testing Strategy
Currently no automated tests. When implementing tests:
- Go: Use standard `testing` package
- Frontend: Consider Vitest (Vite-native) or Jest

### Git Workflow
- **Main branch**: `master`
- Conventional commits preferred
- Recent commits show pattern: `feat:`, `style:` prefixes

## Domain Context
- **File management**: Audio files are the primary domain
- **Security focus**: Path traversal prevention, name sanitization
- **Conflict handling**: Prevents file overwrites with duplicate name checks
- **Simple operations**: Upload, list, rename, delete, and serve files
- **Spanish localization**: Error messages and logs in Spanish

## Important Constraints
- Files must have unique names within the upload directory
- No subdirectory support - flat file structure only
- File names are sanitized to prevent path traversal (no `..`, `/`, `\`)
- Max upload size: 32MB (configured in `ParseMultipartForm`)
- Backend is single-threaded HTTP server (no connection pooling configured)

## External Dependencies
- **Backend**: Zero external dependencies (stdlib only)
- **Frontend**: Minimal dependencies
  - React ecosystem (react, react-dom)
  - Vite for development and building
  - phosphor-react for icons
  - react-mirt for utilities

## File Structure
```
/
├── main.go                    # Go backend server
├── go.mod                     # Go module definition
├── .env.example              # Environment configuration template
├── uploads/                  # File storage directory (created at runtime)
├── frontend/
│   ├── src/
│   │   ├── main.jsx         # React entry point
│   │   ├── App.jsx          # Main app component
│   │   ├── components/      # UI components
│   │   └── services/        # API client layer
│   ├── package.json
│   └── vite.config.js
└── openspec/                # OpenSpec documentation
```
