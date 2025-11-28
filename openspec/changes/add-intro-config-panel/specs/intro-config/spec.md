## ADDED Requirements
### Requirement: Registrar solicitud de intro
El sistema SHALL exponer un endpoint autenticado que permita a un usuario solicitar su sonido de intro usando un archivo existente en `uploads`.

#### Scenario: Solicitud válida se registra
- **WHEN** un usuario autenticado envía `POST /intro` con cuerpo JSON `{"soundName": "<archivo>"}` y el archivo existe en `uploads`
- **THEN** el servidor usa el `user_id` del JWT para registrar en los logs el usuario y `soundName`
- **THEN** responde 200 con un mensaje de confirmación

#### Scenario: Archivo inexistente o inválido
- **WHEN** el usuario envía `POST /intro` con `soundName` vacío o que no existe en `uploads`
- **THEN** el servidor responde con error 400 o 404 indicando que el archivo no es válido

### Requirement: UI para configurar intro
La aplicación SHALL ofrecer un panel lateral con navegación entre “Administrar sonidos” y “Configurar mi intro”, incluyendo selección y previsualización de sonidos existentes.

#### Scenario: Navegación y formulario de intro
- **WHEN** el usuario autenticado abre la app
- **THEN** ve un panel lateral con la opción “Configurar mi intro” además de “Administrar sonidos”
- **THEN** al entrar a “Configurar mi intro” puede buscar/un autocompletado sobre los archivos disponibles, previsualizar un audio y enviar la selección al endpoint de intro
- **THEN** al éxito se muestra un mensaje de confirmación; en caso de error se muestra el mensaje correspondiente
