# Cambio: Panel lateral y configuración de intro

## Por qué
- Necesitamos permitir que cada usuario solicite su sonido de intro sin depender de manejo manual.
- La UI actual solo lista/sube archivos y no hay lugar para pedir la intro.

## Qué cambia
- Agregar un panel lateral de navegación con dos vistas: administrar sonidos (vista actual) y configurar intro.
- Nuevo flujo "Configurar mi intro" donde el usuario autenticado selecciona un sonido existente y envía la solicitud.
- Endpoint backend autenticado que registra en logs la solicitud de intro usando el user_id del JWT; responde 200 en éxito y códigos de error en fallos.

## Impacto
- Backend: rutas protegidas, handler nuevo, lectura de archivos en `uploads`.
- Frontend: App.jsx, nuevos componentes para navegación y formulario de intro, estilos.
