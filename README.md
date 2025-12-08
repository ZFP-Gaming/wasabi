API sencilla en Go para subir y gestionar archivos de audio en una carpeta local (`uploads`) con autenticación mediante Discord OAuth.

## Cómo ejecutar

### Requisitos previos
- Go 1.20+
- Node.js 16+ (para el frontend)
- Cuenta de Discord y una aplicación OAuth configurada

### Configuración de Discord OAuth

1. Ve a [Discord Developer Portal](https://discord.com/developers/applications)
2. Crea una nueva aplicación o selecciona una existente
3. Ve a la sección "OAuth2" en el menú lateral
4. Añade las siguientes URLs de redirección:
   - Desarrollo: `http://localhost:8080/auth/discord/callback`
   - Producción: Tu dominio + `/auth/discord/callback`
5. Copia el **Client ID** y **Client Secret**

### Variables de entorno

Crea un archivo `.env` basado en `.env.example`:

```bash
cp .env.example .env
```

Configura las siguientes variables:

- `PORT`: puerto/addr de escucha (por defecto `8080`, se puede usar `:8080` o `127.0.0.1:8080`)
- `UPLOAD_DIR`: carpeta donde se guardan los archivos (por defecto `uploads`)
- `FRONTEND_ORIGIN`: origen principal del frontend (primer valor si envías varios separados por comas). Se usa para redirecciones OAuth.
- `ALLOWED_ORIGINS`: lista separada por comas de orígenes permitidos para CORS. Incluye automáticamente `FRONTEND_ORIGIN` (ej: `https://wasabi.zfpgaming.cl,http://localhost:5173`)
- `DISCORD_CLIENT_ID`: Client ID de tu aplicación Discord
- `DISCORD_CLIENT_SECRET`: Client Secret de tu aplicación Discord
- `DISCORD_REDIRECT_URI`: URL de callback (ajusta al puerto del backend, ej: `http://localhost:8080/auth/discord/callback`)
- `DISCORD_REQUIRED_GUILD_ID`: ID del servidor de Discord cuyo membership es obligatorio para iniciar sesión (configura el ID del guild)
- `JWT_SECRET`: Clave secreta para firmar tokens JWT (usa una cadena aleatoria larga)

### Ejecución del backend

```bash
go run ./...
```

El servidor escuchará en el puerto configurado (por defecto `http://localhost:8080`).

### Ejecución del frontend

```bash
cd frontend
npm install
npm run dev
```

El frontend se ejecutará en `http://localhost:5173` (Vite default).
Si tu backend escucha en una URL distinta, crea `frontend/.env` con `VITE_API_BASE=<url_del_backend>` para que los botones de login usen el host correcto.

## Endpoints

### Autenticación

- `GET /auth/discord`
  - Inicia el flujo de autenticación OAuth con Discord
  - Redirige al usuario a la página de autorización de Discord

- `GET /auth/discord/callback`
  - Endpoint de callback para Discord OAuth
  - Intercambia el código de autorización por un token de acceso
  - Crea una sesión JWT y establece una cookie httpOnly
  - Redirige al usuario a la aplicación frontend

- `POST /auth/logout`
  - Cierra la sesión del usuario
  - Elimina la cookie de autenticación

- `GET /auth/me` (requiere autenticación)
  - Devuelve información del usuario autenticado
  - Responde con user_id, username, discriminator, y avatar

### Gestión de archivos (requieren autenticación)

- `POST /upload` (multipart/form-data)
  - Campo obligatorio `file`; opcional `filename` para sobrescribir el nombre guardado
  - Devuelve `201` con el nombre final. Rechaza si el nombre ya existe
  - Requiere cookie de autenticación válida

- `GET /files`
  - Lista archivos disponibles en `uploads` con nombre, tamaño y fecha de modificación
  - Requiere autenticación

- `GET /files/{nombre}`
  - Descarga o visualiza el archivo especificado
  - Requiere autenticación

- `PUT /files/{nombreActual}`
  - Cuerpo JSON: `{"newName": "nuevoNombre.ext"}`
  - Cambia el nombre si el archivo existe y no hay conflicto
  - Requiere autenticación

- `DELETE /files/{nombre}`
  - Elimina el archivo especificado
  - Requiere autenticación

## Seguridad

- Todos los endpoints de gestión de archivos requieren autenticación
- Los tokens JWT expiran después de 24 horas
- Las cookies son httpOnly, secure (en producción), y SameSite=Lax
- Los nombres de archivo se sanitizan para prevenir ataques de path traversal
- Protección CSRF mediante validación de estado OAuth
- CORS configurado para permitir credenciales desde el frontend
- Solo los usuarios que pertenecen al servidor de Discord configurado (`DISCORD_REQUIRED_GUILD_ID`) pueden autenticarse y usar los endpoints protegidos

Los nombres se normalizan para evitar rutas peligrosas. Los archivos se guardan en `uploads` (se crea si no existe).
