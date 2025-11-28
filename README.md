API sencilla en Go para subir y gestionar archivos de audio en una carpeta local (`uploads`).

## Cómo ejecutar
- Requiere Go 1.20+.
- Ejecuta el servidor: `go run ./...`
- Escucha en `http://localhost:8080`.

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
