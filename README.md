API sencilla en Go para subir y gestionar archivos de audio en una carpeta local (`uploads`).

## Cómo ejecutar
- Requiere Go 1.20+.
- Variables configurables (via entorno o `.env`):
  - `PORT`: puerto/addr de escucha (por defecto `8080`, se puede usar `:8080` o `127.0.0.1:8080`).
  - `UPLOAD_DIR`: carpeta donde se guardan los archivos (por defecto `uploads`).
- Ejecuta el servidor: `go run ./...`
- Escucha en `http://localhost:8080` si no cambias `PORT`.

## Endpoints
- `POST /upload` (multipart/form-data)
  - Campo obligatorio `file`; opcional `filename` para sobrescribir el nombre guardado.
  - Devuelve `201` con el nombre final. Rechaza si el nombre ya existe.
- `GET /files`
  - Lista archivos disponibles en `uploads` con nombre, tamaño y fecha de modificación.
- `PUT /files/{nombreActual}`
  - Cuerpo JSON: `{"newName": "nuevoNombre.ext"}`.
  - Cambia el nombre si el archivo existe y no hay conflicto.

Los nombres se normalizan para evitar rutas peligrosas. Los archivos se guardan en `uploads` (se crea si no existe).
