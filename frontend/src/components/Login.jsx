import { useAuth } from "../contexts/AuthContext";
import { DiscordLogo } from "phosphor-react";

function Login() {
  const { login, error } = useAuth();

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-header">
          <h1>Wasabi</h1>
          <p className="muted">Gestor de archivos de audio</p>
        </div>

        <div className="login-content">
          <p>Inicia sesión con Discord para acceder a la aplicación</p>

          {error && (
            <div className="alert alert-error">
              {error}
            </div>
          )}

          <button className="discord-button" onClick={login}>
            <DiscordLogo size={24} weight="fill" />
            Iniciar sesión con Discord
          </button>
        </div>

        <div className="login-footer">
          <p className="muted small">
            Usamos Discord para autenticación. No almacenamos contraseñas.
          </p>
        </div>
      </div>
    </div>
  );
}

export default Login;
